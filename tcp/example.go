package tcp

import (
	"fmt"
	"time"
)

// Example 展示如何使用TCP服务
func Example() {
	// 创建配置
	config := &TcpPoolConfig{
		BufferSize:     2048,
		MaxConnections: 100000,
		ConnectTimeout: time.Second * 5,
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 10,
		MaxIdleTime:    time.Minute * 5,
	}

	// 创建TCP服务器
	server, err := NewTCPServer("0.0.0.0:8888", config)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// 设置消息处理函数
	server.SetMessageHandler(func(conn *TcpConnection, msg *TcpMessage) error {
		fmt.Printf("Received message from %s: %s\n", conn.Id, string(msg.Data))

		// 回显消息
		return server.SendTo(conn.Id, []byte(fmt.Sprintf("Echo: %s", msg.Data)))
	})

	// 启动服务器
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

	// 运行10秒后停止
	fmt.Println("TCP server started. Running for 10 seconds...")
	time.Sleep(time.Second * 10)

	// 停止服务器
	if err := server.Stop(); err != nil {
		fmt.Printf("Failed to stop server: %v\n", err)
	}

	fmt.Println("TCP server stopped.")
}

// TestTCP 测试TCP连接
func TestTCP() {
	fmt.Println("=== 测试TCP连接 ===")
	fmt.Println("1. 创建TCP服务器配置")
	config := &TcpPoolConfig{
		BufferSize:     2048,
		MaxConnections: 100000,
		ConnectTimeout: time.Second * 5,
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 10,
		MaxIdleTime:    time.Minute * 5,
	}
	fmt.Println("2. 创建TCP服务器")
	server, err := NewTCPServer("0.0.0.0:8888", config)
	if err != nil {
		fmt.Printf("创建服务器失败：%v\n", err)
		return
	}
	fmt.Println("3. 服务器创建成功")
	fmt.Println("4. 获取在线连接数")
	count := server.Connection.Count()
	fmt.Printf("当前在线连接数：%d\n", count)
	fmt.Println("5. 获取所有在线连接ID")
	connIDs, err := server.GetAllConnIDs()
	if err != nil {
		fmt.Printf("获取在线连接ID失败：%v\n", err)
	} else {
		fmt.Printf("在线连接ID：%v\n", connIDs)
	}
	fmt.Println("6. 启动服务器")
	if err := server.Start(); err != nil {
		fmt.Printf("启动服务器失败：%v\n", err)
		return
	}
	fmt.Println("7. 服务器启动成功，运行2秒后停止")
	time.Sleep(time.Second * 2)
	fmt.Println("8. 停止服务器")
	if err := server.Stop(); err != nil {
		fmt.Printf("停止服务器失败：%v\n", err)
	} else {
		fmt.Println("服务器停止成功")
	}
	fmt.Println("=== TCP测试完成 ===")
}
