package pool

import (
	"time"
)

// ConnType 连接类型
type ConnType string

const (
	ConnTypeWebSocket ConnType = "websocket"
	ConnTypeTCP       ConnType = "tcp"
)

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Type      ConnType  `json:"type" gorm:"index"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"isActive" gorm:"index"`
	LastUsed  time.Time `json:"lastUsed"`
	CreatedAt time.Time `json:"createdAt"`
	// 额外的连接数据，根据不同类型存储不同的信息
	Data map[string]interface{} `json:"data" gorm:"-"`
}
