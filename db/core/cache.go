package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{} // 缓存的数据
	ExpiresAt time.Time   // 过期时间
}

// QueryCache 查询缓存 - 提高重复查询的性能
type QueryCache struct {
	mu       sync.RWMutex          // 读写锁
	items    map[string]*CacheItem // 缓存项
	duration time.Duration         // 默认缓存时长
}

// NewQueryCache 创建查询缓存实例
func NewQueryCache(duration time.Duration) *QueryCache {
	cache := &QueryCache{
		items:    make(map[string]*CacheItem),
		duration: duration,
	}

	// 启动清理协程
	go cache.cleaner()

	return cache
}

// Set 设置缓存
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	qc.items[key] = &CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(qc.duration),
	}
}

// Get 获取缓存
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	item, exists := qc.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

// Delete 删除缓存
func (qc *QueryCache) Delete(key string) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	delete(qc.items, key)
}

// Clear 清空所有缓存
func (qc *QueryCache) Clear() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	qc.items = make(map[string]*CacheItem)
}

// cleaner 定期清理过期缓存
func (qc *QueryCache) cleaner() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		qc.cleanExpired()
	}
}

// cleanExpired 清理过期的缓存项
func (qc *QueryCache) cleanExpired() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	now := time.Now()
	for key, item := range qc.items {
		if now.After(item.ExpiresAt) {
			delete(qc.items, key)
		}
	}
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(sql string, args ...interface{}) string {
	// 将 SQL 和参数组合成字符串
	keyData := sql
	for _, arg := range args {
		keyData += fmt.Sprintf("%v", arg)
	}

	// 计算 MD5 哈希
	hash := md5.Sum([]byte(keyData))
	return hex.EncodeToString(hash[:])
}

// deepCopy 深拷贝数据（使用 JSON 序列化/反序列化）
func deepCopy(src, dst interface{}) error {
	// 序列化为 JSON
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("序列化失败：%w", err)
	}

	// 反序列化到目标
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("反序列化失败：%w", err)
	}

	return nil
}

// WithCache 带缓存的查询装饰器
func (q *QueryBuilder) WithCache(cache *QueryCache) IQuery {
	if cache == nil {
		return q
	}

	// 设置缓存实例
	q.cache = cache
	q.useCache = true

	// 生成缓存键（使用 SQL 和参数）
	sqlStr, args := q.BuildSelect()
	q.cacheKey = GenerateCacheKey(sqlStr, args...)

	return q
}
