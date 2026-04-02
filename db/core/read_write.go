package core

import (
	"database/sql"
	"sync"
	"sync/atomic"
)

// ReadWriteDB 读写分离数据库连接
type ReadWriteDB struct {
	master  *sql.DB      // 主库（写）
	slaves  []*sql.DB    // 从库列表（读）
	policy  ReadPolicy   // 读负载均衡策略
	counter uint64       // 轮询计数器
	mu      sync.RWMutex // 读写锁
}

// NewReadWriteDB 创建读写分离数据库连接
func NewReadWriteDB(master *sql.DB, slaves []*sql.DB, policy ReadPolicy) *ReadWriteDB {
	return &ReadWriteDB{
		master: master,
		slaves: slaves,
		policy: policy,
	}
}

// GetMaster 获取主库连接（用于写操作）
func (rw *ReadWriteDB) GetMaster() *sql.DB {
	return rw.master
}

// GetSlave 获取从库连接（用于读操作）
func (rw *ReadWriteDB) GetSlave() *sql.DB {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if len(rw.slaves) == 0 {
		// 没有从库，使用主库
		return rw.master
	}

	switch rw.policy {
	case Random:
		// 随机选择一个从库
		idx := int(atomic.LoadUint64(&rw.counter)) % len(rw.slaves)
		return rw.slaves[idx]

	case RoundRobin:
		// 轮询选择从库
		idx := int(atomic.AddUint64(&rw.counter, 1)) % len(rw.slaves)
		return rw.slaves[idx]

	case LeastConn:
		// 选择连接数最少的从库（简化实现）
		return rw.selectLeastConn()

	default:
		return rw.slaves[0]
	}
}

// selectLeastConn 选择连接数最少的从库
func (rw *ReadWriteDB) selectLeastConn() *sql.DB {
	if len(rw.slaves) == 0 {
		return rw.master
	}

	minConn := -1
	selected := rw.slaves[0]

	for _, slave := range rw.slaves {
		stats := slave.Stats()
		openConnections := stats.OpenConnections

		if minConn == -1 || openConnections < minConn {
			minConn = openConnections
			selected = slave
		}
	}

	return selected
}

// AddSlave 添加从库
func (rw *ReadWriteDB) AddSlave(slave *sql.DB) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.slaves = append(rw.slaves, slave)
}

// RemoveSlave 移除从库
func (rw *ReadWriteDB) RemoveSlave(slave *sql.DB) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	for i, s := range rw.slaves {
		if s == slave {
			rw.slaves = append(rw.slaves[:i], rw.slaves[i+1:]...)
			break
		}
	}
}

// Close 关闭所有连接
func (rw *ReadWriteDB) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// 关闭主库
	if rw.master != nil {
		if err := rw.master.Close(); err != nil {
			return err
		}
	}

	// 关闭所有从库
	for _, slave := range rw.slaves {
		if err := slave.Close(); err != nil {
			return err
		}
	}

	return nil
}
