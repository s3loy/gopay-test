package repository

import "context"

// TransactionManager provides database transaction support for usecase layer.
// Usecase wraps multi-step operations in a transaction to ensure atomicity.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
