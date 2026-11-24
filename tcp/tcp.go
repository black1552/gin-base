package tcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/black1552/gin-base/pool"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtcp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/os/gtime"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(conn *TcpConnection, msg *TcpMessage) error

// TCPServer TCP服务器结构
type TCPServer struct {
	Address        string
	Config         *TcpPoolConfig
	Listener       *gtcp.Server
	Connection     *ConnectionPool
	Logger         *glog.Logger
	MessageHandler MessageHandler
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// ConnectionPool 连接池结构
type ConnectionPool struct {
	connections map[string]*TcpConnection
	sqlitePool  *pool.SQLitePool
	mutex       sync.RWMutex
	config      *TcpPoolConfig
	logger      *glog.Logger
}

// NewTCPServer 创建一个新的TCP服务器
func NewTCPServer(address string, config *TcpPoolConfig) (*TCPServer, error) {
	logger := g.Log(address)
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化SQLite连接池
	sqlitePool, err := pool.NewSQLitePool()
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite connPool: %w", err)
	}

	connPool := &ConnectionPool{
		connections: make(map[string]*TcpConnection),
		sqlitePool:  sqlitePool,
		config:      config,
		logger:      logger,
	}

	server := &TCPServer{
		Address:    address,
		Config:     config,
		Connection: connPool,
		Logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	server.Listener = gtcp.NewServer(address, server.handleConnection)
	return server, nil
}

// SetMessageHandler 设置消息处理函数
func (s *TCPServer) SetMessageHandler(handler MessageHandler) {
	s.MessageHandler = handler
}

// Start 启动TCP服务器
func (s *TCPServer) Start() error {
	s.Logger.Info(s.ctx, fmt.Sprintf("TCP server starting on %s", s.Address))
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		if err := s.Listener.Run(); err != nil {
			s.Logger.Error(s.ctx, fmt.Sprintf("TCP server stopped with error: %v", err))
		}
	}()
	return nil
}

// Stop 停止TCP服务器
func (s *TCPServer) Stop() error {
	s.Logger.Info(s.ctx, "TCP server stopping...")
	s.cancel()
	s.Listener.Close()
	s.wg.Wait()
	s.Connection.Clear()
	// 关闭SQLite连接池
	if err := s.Connection.sqlitePool.Close(); err != nil {
		s.Logger.Error(s.ctx, fmt.Sprintf("Failed to close SQLite pool: %v", err))
		// 不影响服务器停止，仅记录错误
	}
	s.Logger.Info(s.ctx, "TCP server stopped")
	return nil
}

// handleConnection 处理新连接
func (s *TCPServer) handleConnection(conn *gtcp.Conn) {
	// 生成连接ID
	connID := fmt.Sprintf("%s_%d", conn.RemoteAddr().String(), gtime.TimestampNano())

	// 创建连接对象
	tcpConn := &TcpConnection{
		Id:        connID,
		Address:   conn.RemoteAddr().String(),
		Server:    *conn,
		IsActive:  true,
		LastUsed:  time.Now(),
		CreatedAt: time.Now(),
	}

	// 添加到连接池
	s.Connection.Add(tcpConn)
	s.Logger.Info(s.ctx, fmt.Sprintf("New connection established: %s", connID))

	// 存储到SQLite
	connInfo := &pool.ConnectionInfo{
		ID:        connID,
		Type:      pool.ConnTypeTCP,
		Address:   conn.RemoteAddr().String(),
		IsActive:  true,
		LastUsed:  time.Now(),
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"localAddress": conn.LocalAddr().String(),
		},
	}
	if err := s.Connection.sqlitePool.Add(connInfo); err != nil {
		s.Logger.Error(s.ctx, fmt.Sprintf("Failed to store connection to SQLite: %v", err))
		// 不影响连接建立，仅记录错误
	}

	// 启动消息接收协程
	go s.receiveMessages(tcpConn)
}

// receiveMessages 接收消息
func (s *TCPServer) receiveMessages(conn *TcpConnection) {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Error(s.ctx, fmt.Sprintf("Panic in receiveMessages: %v", err))
		}
		s.Connection.Remove(conn.Id)
		conn.Server.Close()
		// 从SQLite移除
		if err := s.Connection.sqlitePool.Remove(conn.Id); err != nil {
			s.Logger.Error(s.ctx, fmt.Sprintf("Failed to remove connection from SQLite: %v", err))
			// 不影响连接关闭，仅记录错误
		}
		s.Logger.Info(s.ctx, fmt.Sprintf("Connection closed: %s", conn.Id))
	}()

	buffer := make([]byte, s.Config.BufferSize)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 设置读取超时
			conn.Server.SetReadDeadline(time.Now().Add(s.Config.ReadTimeout))

			// 读取数据
			n, err := conn.Server.Read(buffer)
			if err != nil {
				s.Logger.Error(s.ctx, fmt.Sprintf("Read error from %s: %v", conn.Id, err))
				return
			}

			if n > 0 {
				// 更新最后使用时间
				now := time.Now()
				conn.Mutex.Lock()
				conn.LastUsed = now
				// 将读取的数据添加到连接的缓冲区
				conn.buffer = append(conn.buffer, buffer[:n]...)
				conn.Mutex.Unlock()

				// 更新SQLite中的连接信息
				connInfo, err := s.Connection.sqlitePool.Get(conn.Id)
				if err == nil && connInfo != nil {
					connInfo.LastUsed = now
					if err := s.Connection.sqlitePool.Update(connInfo); err != nil {
						s.Logger.Error(s.ctx, fmt.Sprintf("Failed to update connection in SQLite: %v", err))
						// 不影响消息处理，仅记录错误
					}
				}

				// 解析消息帧
				s.parseMessageFrames(conn)
			}
		}
	}
}

// parseMessageFrames 解析消息帧
func (s *TCPServer) parseMessageFrames(conn *TcpConnection) {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	for {
		// 检查缓冲区是否有足够的数据来读取长度前缀
		if len(conn.buffer) < messageLengthPrefixSize {
			// 数据不足，等待下一次读取
			return
		}

		// 读取长度前缀
		length := binary.BigEndian.Uint32(conn.buffer[:messageLengthPrefixSize])

		// 检查缓冲区是否有足够的数据来读取完整的消息
		if len(conn.buffer) < messageLengthPrefixSize+int(length) {
			// 数据不足，等待下一次读取
			return
		}

		// 提取消息数据
		data := conn.buffer[messageLengthPrefixSize : messageLengthPrefixSize+int(length)]

		// 移除已处理的消息数据
		conn.buffer = conn.buffer[messageLengthPrefixSize+int(length):]

		// 创建消息对象
		msg := &TcpMessage{
			Id:        fmt.Sprintf("msg_%d", gtime.TimestampNano()),
			ConnId:    conn.Id,
			Data:      data,
			Timestamp: time.Now(),
			IsSend:    false,
		}

		// 使用协程池处理消息，避免阻塞
		grpool.AddWithRecover(s.ctx, func(ctx context.Context) {
			if s.MessageHandler != nil {
				if err := s.MessageHandler(conn, msg); err != nil {
					s.Logger.Error(s.ctx, fmt.Sprintf("Message handling error: %v", err))
				}
			}
		}, func(ctx context.Context, err error) {
			s.Logger.Error(ctx, fmt.Sprintf("Message handling error: %v", err))
		})
	}
}

// SendTo 发送消息到指定连接
func (s *TCPServer) SendTo(connID string, data []byte) error {
	conn := s.Connection.Get(connID)
	if conn == nil {
		return fmt.Errorf("connection not found: %s", connID)
	}
	return s.sendMessage(conn, data)
}

// SendToAll 发送消息到所有连接
func (s *TCPServer) SendToAll(data []byte) error {
	conns := s.Connection.GetAll()
	for _, conn := range conns {
		if err := s.sendMessage(conn, data); err != nil {
			s.Logger.Error(s.ctx, fmt.Sprintf("Send to %s failed: %v", conn.Id, err))
			// 继续发送给其他连接
		}
	}
	return nil
}

// sendMessage 发送消息
func (s *TCPServer) sendMessage(conn *TcpConnection, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	// 设置写入超时
	conn.Server.SetWriteDeadline(time.Now().Add(s.Config.WriteTimeout))

	// 创建消息帧：4字节长度前缀 + 消息数据
	frame := make([]byte, messageLengthPrefixSize+len(data))
	// 写入长度前缀（大端序）
	binary.BigEndian.PutUint32(frame[:messageLengthPrefixSize], uint32(len(data)))
	// 写入消息数据
	copy(frame[messageLengthPrefixSize:], data)

	// 发送数据
	_, err := conn.Server.Write(frame)
	if err != nil {
		return err
	}

	// 更新最后使用时间
	now := time.Now()
	conn.LastUsed = now

	// 更新SQLite中的连接信息
	connInfo, err := s.Connection.sqlitePool.Get(conn.Id)
	if err == nil && connInfo != nil {
		connInfo.LastUsed = now
		if err := s.Connection.sqlitePool.Update(connInfo); err != nil {
			s.Logger.Error(s.ctx, fmt.Sprintf("Failed to update connection in SQLite: %v", err))
			// 不影响消息发送，仅记录错误
		}
	}

	return nil
}

// Kick 强制退出客户端
func (s *TCPServer) Kick(connID string) error {
	conn := s.Connection.Get(connID)
	if conn == nil {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 关闭连接
	conn.Server.Close()
	// 从连接池移除
	s.Connection.Remove(connID)
	// 从SQLite移除
	if err := s.Connection.sqlitePool.Remove(connID); err != nil {
		s.Logger.Error(s.ctx, fmt.Sprintf("Failed to remove connection from SQLite: %v", err))
		// 不影响连接关闭，仅记录错误
	}

	s.Logger.Info(s.ctx, fmt.Sprintf("Kicked connection: %s", connID))
	return nil
}

// GetAllConnIDs 获取所有在线连接的ID列表
func (s *TCPServer) GetAllConnIDs() ([]string, error) {
	return s.Connection.GetAllConnIDs()
}

// Add 添加连接到连接池
func (p *ConnectionPool) Add(conn *TcpConnection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.connections[conn.Id] = conn
}

// Get 获取连接
func (p *ConnectionPool) Get(connID string) *TcpConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.connections[connID]
}

// GetAll 获取所有连接
func (p *ConnectionPool) GetAll() []*TcpConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	conns := make([]*TcpConnection, 0, len(p.connections))
	for _, conn := range p.connections {
		conns = append(conns, conn)
	}
	return conns
}

// Remove 从连接池移除连接
func (p *ConnectionPool) Remove(connID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.connections, connID)
}

// Clear 清空连接池
func (p *ConnectionPool) Clear() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for connID, conn := range p.connections {
		conn.Server.Close()
		delete(p.connections, connID)
	}
}

// Count 获取连接数量
func (p *ConnectionPool) Count() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.connections)
}

// GetAllConnIDs 获取所有在线连接的ID列表
func (p *ConnectionPool) GetAllConnIDs() ([]string, error) {
	return p.sqlitePool.GetAllConnIDs()
}
