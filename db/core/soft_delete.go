package core

import (
	"time"
)

// SoftDelete 软删除模型 - 嵌入到需要软删除的模型中
type SoftDelete struct {
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"` // 删除时间（为空表示未删除）
}

// IsDeleted 检查是否已删除
func (sd *SoftDelete) IsDeleted() bool {
	return sd.DeletedAt != nil
}

// Delete 标记为已删除
func (sd *SoftDelete) Delete() {
	now := time.Now()
	sd.DeletedAt = &now
}

// Restore 恢复（取消删除）
func (sd *SoftDelete) Restore() {
	sd.DeletedAt = nil
}

// ISoftDeleter 软删除接口 - 定义软删除相关方法
type ISoftDeleter interface {
	IsDeleted() bool
	Delete()
	Restore()
}

// applySoftDelete 在查询中应用软删除过滤
func applySoftDelete(q IQuery, unscoped bool) IQuery {
	if unscoped {
		// 忽略软删除，包含已删除的记录
		return q
	}

	// 默认只查询未删除的记录
	return q.Where("deleted_at IS NULL")
}
