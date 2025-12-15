package database

import (
	"context"

	"gorm.io/gorm"
)

// TransactionManagerI is the interface for transaction management
type TransactionManagerI interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManagerI {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = WithTx(ctx, tx)
		return fn(ctx)
	})
}
