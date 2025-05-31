package snowdrop

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// TransactionManager provides an interface for managing database transactions.
// It allows creating new transactions and retrieving currently active transactions
// within a given context.
//
// The interface is designed to abstract transaction handling and provide
// consistent transaction management across different database operations.
type TransactionManager interface {
	// NewTransaction creates and returns a new database transaction with specified options.
	// The transaction's lifecycle is managed within the provided context.
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (*gorm.DB, error)

	// GetCurrentTx returns the currently active transaction associated with the given context.
	// Returns nil if no transaction is currently active in the context.
	GetCurrentTx(ctx context.Context) *gorm.DB
}

// TxDataFunc defines the function signature for transactional operations.
type TxDataFunc[T any] func(ctx context.Context, tx *gorm.DB) (T, error)

// TxFunc defines a transactional function that only returns an error.
type TxFunc func(ctx context.Context, tx *gorm.DB) error
