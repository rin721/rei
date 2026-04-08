package dbtx

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type txContextKey struct{}

// Manager 提供带上下文传播的事务执行能力。
type Manager struct {
	db *gorm.DB
}

// New 创建一个新的事务管理器。
func New(db *gorm.DB) (*Manager, error) {
	if db == nil {
		return nil, errors.New("gorm db is required")
	}

	return &Manager{db: db}, nil
}

// WithTx 在默认事务选项下执行回调。
func (m *Manager) WithTx(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return m.WithTxOptions(ctx, nil, fn)
}

// WithTxOptions 使用指定事务选项执行回调。
func (m *Manager) WithTxOptions(ctx context.Context, options *sql.TxOptions, fn func(context.Context, *gorm.DB) error) error {
	if fn == nil {
		return errors.New("transaction callback is required")
	}

	tx := m.db.WithContext(ctx).Begin(options)
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetDB 从上下文中优先获取事务实例，否则返回根数据库。
func (m *Manager) GetDB(ctx context.Context) *gorm.DB {
	if ctx != nil {
		if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
			return tx
		}
	}

	return m.db
}
