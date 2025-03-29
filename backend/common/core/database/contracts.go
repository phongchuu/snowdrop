package database

import (
	"context"
	"database/sql"
)

type transactionManagerType string

type Database interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TransactionManager interface {
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetCurrentTx(ctx context.Context) *sql.Tx
}

// TxDataFunc defines the function signature for transactional operations.
type TxDataFunc[T any] func(ctx context.Context, tx *sql.Tx) (T, error)

// TxFunc defines a transactional function that only returns an error.
type TxFunc func(ctx context.Context, tx *sql.Tx) error
