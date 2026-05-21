package postgresql

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/repository"
	"gorm.io/gorm"
)

type txKey struct{}

// WithTx injects a GORM transaction into the context for repository methods to use.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) repository.TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}
