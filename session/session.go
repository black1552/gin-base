package session

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionManager Session管理器，负责Session的创建、获取和销毁
type SessionManager struct {
	sessions map[string]*Session // 存储所有活跃的Session，key为SessionID
	mutex    sync.RWMutex        // 读写锁，保证并发安全
	maxAge   time.Duration       // Session最大存活时间
}

// Session 单个Session对象，用于存储用户数据
type Session struct {
	ID         string                 // Session唯一标识符
	Data       map[string]interface{} // 存储的数据，键值对形式
	CreateTime time.Time              // 创建时间
	LastAccess time.Time              // 最后访问时间
	mutex      sync.RWMutex           // 读写锁，保证并发安全
}

var (
	// defaultManager 默认的Session管理器实例
	defaultManager *SessionManager
	// once 确保只初始化一次
	once sync.Once
)

// init 包初始化时自动创建默认管理器
func init() {
	InitDefaultManager(30 * time.Minute) // 默认30分钟过期
}

// InitDefaultManager 初始化默认Session管理器
// 参数 duration: Session的最大存活时间
func InitDefaultManager(duration time.Duration) {
	once.Do(func() {
		defaultManager = NewSessionManager(duration)
	})
}

// NewSessionManager 创建一个新的Session管理器
// 参数 duration: Session的最大存活时间
// 返回值: Session管理器指针
func NewSessionManager(duration time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		maxAge:   duration,
	}

	// 启动定时清理任务，每5分钟清理一次过期的Session
	go sm.startCleanupTicker(5 * time.Minute)

	return sm
}

// startCleanupTicker 启动定时清理过期Session的任务
// 参数 interval: 清理间隔时间
func (sm *SessionManager) startCleanupTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		sm.CleanupExpiredSessions()
	}
}

// CreateSession 创建新的Session并返回Session ID
// 参数 c: Gin上下文对象
// 返回值: Session ID字符串
func (sm *SessionManager) CreateSession(c *gin.Context) string {
	sessionID := uuid.New().String()

	session := &Session{
		ID:         sessionID,
		Data:       make(map[string]interface{}),
		CreateTime: time.Now(),
		LastAccess: time.Now(),
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()

	// 设置Cookie，保存Session ID
	c.SetCookie(
		"session_id",             // Cookie名称
		sessionID,                // Cookie值
		int(sm.maxAge.Seconds()), // 过期时间（秒）
		"/",                      // 路径
		"",                       // 域名（空表示当前域名）
		false,                    // 是否仅HTTPS
		true,                     // 是否HttpOnly（防止XSS攻击）
	)

	return sessionID
}

// GetSession 根据Gin上下文获取Session对象
// 参数 c: Gin上下文对象
// 返回值: Session对象指针，如果不存在则返回nil
func (sm *SessionManager) GetSession(c *gin.Context) *Session {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return nil
	}

	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil
	}

	// 检查Session是否过期
	if time.Since(session.LastAccess) > sm.maxAge {
		sm.DestroySession(sessionID)
		return nil
	}

	// 更新最后访问时间
	session.mutex.Lock()
	session.LastAccess = time.Now()
	session.mutex.Unlock()

	return session
}

// GetSessionByID 根据Session ID获取Session对象
// 参数 sessionID: Session的唯一标识符
// 返回值: Session对象指针，如果不存在则返回nil
func (sm *SessionManager) GetSessionByID(sessionID string) *Session {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil
	}

	// 检查Session是否过期
	if time.Since(session.LastAccess) > sm.maxAge {
		sm.DestroySession(sessionID)
		return nil
	}

	// 更新最后访问时间
	session.mutex.Lock()
	session.LastAccess = time.Now()
	session.mutex.Unlock()

	return session
}

// DestroySession 销毁指定的Session
// 参数 sessionID: Session的唯一标识符
func (sm *SessionManager) DestroySession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()
}

// DestroySessionByContext 根据Gin上下文销毁Session
// 参数 c: Gin上下文对象
func (sm *SessionManager) DestroySessionByContext(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return
	}

	sm.DestroySession(sessionID)

	// 清除Cookie
	c.SetCookie(
		"session_id",
		"",
		-1, // 立即过期
		"/",
		"",
		false,
		true,
	)
}

// CleanupExpiredSessions 清理所有过期的Session
func (sm *SessionManager) CleanupExpiredSessions() {
	now := time.Now()
	expiredIDs := make([]string, 0)

	sm.mutex.RLock()
	for id, session := range sm.sessions {
		session.mutex.RLock()
		if now.Sub(session.LastAccess) > sm.maxAge {
			expiredIDs = append(expiredIDs, id)
		}
		session.mutex.RUnlock()
	}
	sm.mutex.RUnlock()

	// 删除过期的Session
	sm.mutex.Lock()
	for _, id := range expiredIDs {
		delete(sm.sessions, id)
	}
	sm.mutex.Unlock()
}

// GetSessionCount 获取当前活跃的Session数量
// 返回值: Session数量
func (sm *SessionManager) GetSessionCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.sessions)
}

// Set 在Session中设置键值对
// 参数 key: 键名
// 参数 value: 值
func (s *Session) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data[key] = value
}

// Get 从Session中获取值
// 参数 key: 键名
// 返回值: 对应的值，如果不存在则返回nil
func (s *Session) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Data[key]
}

// GetString 从Session中获取字符串类型的值
// 参数 key: 键名
// 返回值: 字符串值，如果不存在或类型不匹配则返回空字符串
func (s *Session) GetString(key string) string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if val, ok := s.Data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 从Session中获取整数类型的值
// 参数 key: 键名
// 返回值: 整数值，如果不存在或类型不匹配则返回0
func (s *Session) GetInt(key string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if val, ok := s.Data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64: // JSON解码时数字会变成float64
			return int(v)
		}
	}
	return 0
}

// GetFloat64 从Session中获取浮点数类型的值
// 参数 key: 键名
// 返回值: 浮点数值，如果不存在或类型不匹配则返回0
func (s *Session) GetFloat64(key string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if val, ok := s.Data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return 0
}

// GetBool 从Session中获取布尔类型的值
// 参数 key: 键名
// 返回值: 布尔值，如果不存在或类型不匹配则返回false
func (s *Session) GetBool(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if val, ok := s.Data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Delete 从Session中删除指定的键
// 参数 key: 要删除的键名
func (s *Session) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.Data, key)
}

// Clear 清空Session中的所有数据
func (s *Session) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data = make(map[string]interface{})
}

// Has 检查Session中是否存在指定的键
// 参数 key: 键名
// 返回值: 如果存在返回true，否则返回false
func (s *Session) Has(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.Data[key]
	return exists
}

// GetAll 获取Session中的所有数据
// 返回值: 包含所有数据的map副本
func (s *Session) GetAll() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回副本，避免外部修改
	result := make(map[string]interface{}, len(s.Data))
	for k, v := range s.Data {
		result[k] = v
	}
	return result
}

// ToJSON 将Session数据转换为JSON字符串
// 返回值: JSON字符串，如果转换失败则返回错误
func (s *Session) ToJSON() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.Marshal(s.Data)
	if err != nil {
		return "", fmt.Errorf("序列化Session数据失败: %w", err)
	}

	return string(data), nil
}

// FromJSON 从JSON字符串恢复Session数据
// 参数 jsonStr: JSON字符串
// 返回值: 如果解析失败则返回错误
func (s *Session) FromJSON(jsonStr string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return fmt.Errorf("反序列化Session数据失败: %w", err)
	}

	s.Data = data
	return nil
}

// GetDefaultManager 获取默认的Session管理器
// 返回值: 默认Session管理器指针
func GetDefaultManager() *SessionManager {
	return defaultManager
}

// CreateSession 使用默认管理器创建Session
// 参数 c: Gin上下文对象
// 返回值: Session ID字符串
func CreateSession(c *gin.Context) string {
	return defaultManager.CreateSession(c)
}

// GetSession 使用默认管理器获取Session
// 参数 c: Gin上下文对象
// 返回值: Session对象指针
func GetSession(c *gin.Context) *Session {
	return defaultManager.GetSession(c)
}

// DestroySession 使用默认管理器销毁Session
// 参数 c: Gin上下文对象
func DestroySession(c *gin.Context) {
	defaultManager.DestroySessionByContext(c)
}
