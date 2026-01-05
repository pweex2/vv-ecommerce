package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// TransactionManager 定义了管理数据库事务的接口
type TransactionManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// GormTransactionManager 是基于 GORM 的实现
type GormTransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Transaction 开启一个事务，并将事务对象注入到 context 中
func (m *GormTransactionManager) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将 tx 注入到 context 中
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// GetDB 尝试从 context 中获取事务 DB，如果不存在则返回默认 DB
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
