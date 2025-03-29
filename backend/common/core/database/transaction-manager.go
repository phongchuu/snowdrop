package database

import (
	"context"
	"database/sql"
	"fmt"
)

type TransactionManagerImpl struct {
	db *sql.DB
}

// NewTransactionManager constructs a new TransactionManagerImpl using the provided database connection.
func NewTransactionManager(db *sql.DB) *TransactionManagerImpl {
	return &TransactionManagerImpl{db: db}
}

func (m TransactionManagerImpl) configureTransaction(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`SELECT set_config('session.requester', $1, true)
        WHERE current_setting('session.requester', true) IS DISTINCT FROM $1;`,
		"xxx",
	)
	if err != nil {
		return fmt.Errorf("failed to set session.requester: %w", err)
	}

	return nil
}

func (m TransactionManagerImpl) NewTransaction(
	ctx context.Context,
	opts *sql.TxOptions,
) (*sql.Tx, error) {
	tx, err := m.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := m.configureTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to set session.requester: %w", err)
	}

	return tx, nil
}

func (m TransactionManagerImpl) GetCurrentTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(transactionManagerCtxKey).(*sql.Tx); ok {
		return tx
	}

	return nil
}

// executeTx executes a transactional operation by running the specified callback within a database transaction.
// It first checks if a transaction is already available in the context; if so, it uses the existing transaction.
// Otherwise, it begins a new transaction with the provided options and attaches it to the context.
// The function ensures that the transaction is committed on success or rolled back if an error occurs or a panic is recovered,
// returning either the result of the callback or an error if any part of the transaction process fails.
func executeTx[T any](
	ctx context.Context,
	transactionManager TransactionManager,
	opts *sql.TxOptions,
	fn TxDataFunc[T],
) (T, error) {
	if existingTx := transactionManager.GetCurrentTx(ctx); existingTx != nil {
		return fn(ctx, existingTx)
	}

	tx, err := transactionManager.NewTransaction(ctx, opts)
	if err != nil {
		return *new(T), fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, transactionManagerCtxKey, tx)

	var result T

	var operationErr error

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()

			panic(p)
		}

		if operationErr == nil {
			if commitErr := tx.Commit(); commitErr != nil {
				operationErr = fmt.Errorf("commit failed: %w", commitErr)
			}
		}
	}()

	result, operationErr = fn(txCtx, tx)
	if operationErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return *new(T), fmt.Errorf(
				"operation failed: %w, rollback failed: %w",
				operationErr,
				rbErr,
			)
		}

		return *new(T), operationErr
	}

	return result, nil
}

// RunTx executes the provided transaction function within a managed transaction context using default options.
// It wraps the function to run within a transaction, ensuring that a new transaction is started if one does not already exist.
// The function commits the transaction on success or rolls it back if an error occurs, returning any encountered error.
func RunTx(
	ctx context.Context,
	transactionManager TransactionManager,
	fn TxFunc,
) error {
	_, err := executeTx(
		ctx,
		transactionManager,
		nil,
		func(ctx context.Context, tx *sql.Tx) (any, error) {
			return nil, fn(ctx, tx)
		},
	)

	return err
}

// RunTxWithOptions executes the provided function within a new transaction configured with the specified options.
// It begins a transaction using the given *sql.TxOptions and delegates execution to a transactional context that
// commits on success or rolls back if an error is encountered.
func RunTxWithOptions(
	ctx context.Context,
	transactionManager TransactionManager,
	opts *sql.TxOptions,
	fn TxFunc,
) error {
	_, err := executeTx(
		ctx,
		transactionManager,
		opts,
		func(ctx context.Context, tx *sql.Tx) (any, error) {
			return nil, fn(ctx, tx)
		},
	)

	return err
}

// RunTxWithData executes the provided function within a new transaction and returns its result of type T.
// It wraps executeTx with no explicit transaction options, automatically handling commit and rollback.
func RunTxWithData[T any](
	ctx context.Context,
	transactionManager TransactionManager,
	fn TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, nil, fn)
}

// RunTxWithDataAndOptions executes a data-returning function within a managed transaction using the provided SQL transaction options.
// It ensures that the transactional function is run in a context where the transaction is committed on success or rolled back on failure,
// returning the result of type T along with any error encountered during the transaction lifecycle.
func RunTxWithDataAndOptions[T any](
	ctx context.Context,
	transactionManager TransactionManager,
	opts *sql.TxOptions,
	fn TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, opts, fn)
}
