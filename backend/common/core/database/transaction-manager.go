package database

import (
	"context"
	"database/sql"
	"fmt"
)

type TransactionManagerImpl struct {
	db *sql.DB
}

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

// executeTx is the core transaction handler.
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

func RunTxWithData[T any](
	ctx context.Context,
	transactionManager TransactionManager,
	fn TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, nil, fn)
}

func RunTxWithDataAndOptions[T any](
	ctx context.Context,
	transactionManager TransactionManager,
	opts *sql.TxOptions,
	fn TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, opts, fn)
}
