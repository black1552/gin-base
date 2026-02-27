package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.magicany.cc/black1552/gf-common/db"
	"gorm.io/gorm"
)

// SQLitePool SQLite连接池
type SQLitePool struct {
	db     *gorm.DB
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	// 内存缓存，提高并发性能
	cache map[string]*ConnectionInfo
}

// NewSQLitePool 创建SQLite连接池
func NewSQLitePool() (*SQLitePool, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 检查数据库连接是否正常
	if db.Db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	// 自动迁移ConnectionInfo模型
	err := db.Db.AutoMigrate(&ConnectionInfo{})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to migrate connection info model: %w", err)
	}

	return &SQLitePool{
		db:     db.Db,
		ctx:    ctx,
		cancel: cancel,
		cache:  make(map[string]*ConnectionInfo),
	}, nil
}

// Close 关闭连接池
func (p *SQLitePool) Close() error {
	p.cancel()
	// SQLite连接由db包管理，不需要在这里关闭
	return nil
}

// Add 添加连接
func (p *SQLitePool) Add(conn *ConnectionInfo) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 存储到SQLite
	result := p.db.Create(conn)
	if result.Error != nil {
		return fmt.Errorf("failed to store connection: %w", result.Error)
	}

	// 更新内存缓存
	p.cache[conn.ID] = conn
	return nil
}

// Get 获取连接
func (p *SQLitePool) Get(connID string) (*ConnectionInfo, error) {
	p.mutex.RLock()
	// 先从内存缓存获取
	if conn, ok := p.cache[connID]; ok {
		p.mutex.RUnlock()
		return conn, nil
	}
	p.mutex.RUnlock()

	// 从SQLite获取
	var connInfo ConnectionInfo
	result := p.db.First(&connInfo, "id = ?", connID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connection: %w", result.Error)
	}

	// 更新内存缓存
	p.mutex.Lock()
	p.cache[connID] = &connInfo
	p.mutex.Unlock()

	return &connInfo, nil
}

// Remove 移除连接
func (p *SQLitePool) Remove(connID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 从SQLite删除
	result := p.db.Delete(&ConnectionInfo{}, "id = ?", connID)
	if result.Error != nil {
		return fmt.Errorf("failed to remove connection: %w", result.Error)
	}

	// 从内存缓存删除
	delete(p.cache, connID)
	return nil
}

// Update 更新连接信息
func (p *SQLitePool) Update(conn *ConnectionInfo) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 存储到SQLite
	result := p.db.Save(conn)
	if result.Error != nil {
		return fmt.Errorf("failed to update connection: %w", result.Error)
	}

	// 更新内存缓存
	p.cache[conn.ID] = conn
	return nil
}

// GetAll 获取所有连接
func (p *SQLitePool) GetAll() ([]*ConnectionInfo, error) {
	p.mutex.RLock()
	// 如果内存缓存不为空，直接返回缓存
	if len(p.cache) > 0 {
		conns := make([]*ConnectionInfo, 0, len(p.cache))
		for _, conn := range p.cache {
			conns = append(conns, conn)
		}
		p.mutex.RUnlock()
		return conns, nil
	}
	p.mutex.RUnlock()

	// 从SQLite获取所有连接
	var conns []*ConnectionInfo
	result := p.db.Find(&conns)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all connections: %w", result.Error)
	}

	// 更新内存缓存
	p.mutex.Lock()
	for _, conn := range conns {
		p.cache[conn.ID] = conn
	}
	p.mutex.Unlock()

	return conns, nil
}

// GetByType 根据类型获取连接
func (p *SQLitePool) GetByType(connType ConnType) ([]*ConnectionInfo, error) {
	allConns, err := p.GetAll()
	if err != nil {
		return nil, err
	}

	var filtered []*ConnectionInfo
	for _, conn := range allConns {
		if conn.Type == connType {
			filtered = append(filtered, conn)
		}
	}

	return filtered, nil
}

// Count 获取连接数量
func (p *SQLitePool) Count() (int, error) {
	p.mutex.RLock()
	// 如果内存缓存不为空，直接返回缓存大小
	if len(p.cache) > 0 {
		count := len(p.cache)
		p.mutex.RUnlock()
		return count, nil
	}
	p.mutex.RUnlock()

	// 从SQLite统计数量
	var count int64
	result := p.db.Model(&ConnectionInfo{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count connections: %w", result.Error)
	}

	return int(count), nil
}

// GetAllConnIDs 获取所有在线连接的ID列表
func (p *SQLitePool) GetAllConnIDs() ([]string, error) {
	p.mutex.RLock()
	// 如果内存缓存不为空，从缓存中提取在线连接的ID
	if len(p.cache) > 0 {
		ids := make([]string, 0, len(p.cache))
		for id, conn := range p.cache {
			if conn.IsActive {
				ids = append(ids, id)
			}
		}
		p.mutex.RUnlock()
		return ids, nil
	}
	p.mutex.RUnlock()

	// 从SQLite获取所有在线连接的ID
	var conns []*ConnectionInfo
	result := p.db.Where("is_active = ?", true).Find(&conns)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all connection IDs: %w", result.Error)
	}

	ids := make([]string, 0, len(conns))
	for _, conn := range conns {
		ids = append(ids, conn.ID)
	}

	return ids, nil
}

// CleanupInactive 清理不活跃的连接
func (p *SQLitePool) CleanupInactive(duration time.Duration) error {
	allConns, err := p.GetAll()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, conn := range allConns {
		if !conn.IsActive || now.Sub(conn.LastUsed) > duration {
			if err := p.Remove(conn.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
