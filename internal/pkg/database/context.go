package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func WithTx(ctx context.Context, txTx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, txTx)
}

func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
