package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	snowdrop "internal.snowdrop/framework"
)

type transactionManagerType string

type TransactionManagerImpl struct {
	db *gorm.DB
}

const transactionManagerCtxKey transactionManagerType = "tx"

var _ snowdrop.TransactionManager = (*TransactionManagerImpl)(nil)

func NewTransactionManager(db *gorm.DB) *TransactionManagerImpl {
	return &TransactionManagerImpl{db: db}
}

func (m TransactionManagerImpl) NewTransaction(
	ctx context.Context,
	opts *sql.TxOptions,
) (*gorm.DB, error) {
	tx := m.db.WithContext(ctx).Begin(opts)

	if err := m.configureTransaction(ctx, tx); err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("failed to set session.requester: %w", err)
	}

	return tx, nil
}

func (TransactionManagerImpl) GetCurrentTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transactionManagerCtxKey).(*gorm.DB); ok {
		return tx
	}

	return nil
}

func (TransactionManagerImpl) configureTransaction(ctx context.Context, tx *gorm.DB) error {
	currentSession, err := snowdrop.GetCurrentSessionFromCtx(ctx)
	if err != nil || errors.Is(err, snowdrop.ErrNoSession) {
		err = tx.Exec(
			`SELECT set_config('session.requester', u.id::VARCHAR(255), true)
            FROM public.users u
            WHERE u.username = $1;`,
			"system",
		).Error
	}

	if currentSession != nil {
		if !currentSession.UserID.Valid {
			err = tx.Exec(
				`SELECT set_config('session.requester', u.id::VARCHAR(255), true)
                FROM public.users u
                WHERE u.username = $1;`,
				"annonymous",
			).Error
		}

		if currentSession.UserID.Valid {
			err = tx.Exec(
				`SELECT set_config('session.requester', $1, true);`,
				currentSession.UserID.UUID.String(),
			).Error
		}
	}

	if err != nil {
		return fmt.Errorf("failed to set session.requester: %w", err)
	}

	return nil
}

// executeTx is the core transaction handler.
func executeTx[T any](
	ctx context.Context,
	transactionManager snowdrop.TransactionManager,
	opts *sql.TxOptions,
	fn snowdrop.TxDataFunc[T],
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
	transactionManager snowdrop.TransactionManager,
	fn snowdrop.TxFunc,
) error {
	_, err := executeTx(
		ctx,
		transactionManager,
		nil,
		func(ctx context.Context, tx *gorm.DB) (any, error) {
			return nil, fn(ctx, tx)
		},
	)

	return err
}

func RunTxWithOptions(
	ctx context.Context,
	transactionManager snowdrop.TransactionManager,
	opts *sql.TxOptions,
	fn snowdrop.TxFunc,
) error {
	_, err := executeTx(
		ctx,
		transactionManager,
		opts,
		func(ctx context.Context, tx *gorm.DB) (any, error) {
			return nil, fn(ctx, tx)
		},
	)

	return err
}

func RunTxWithData[T any](
	ctx context.Context,
	transactionManager snowdrop.TransactionManager,
	fn snowdrop.TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, nil, fn)
}

func RunTxWithDataAndOptions[T any](
	ctx context.Context,
	transactionManager snowdrop.TransactionManager,
	opts *sql.TxOptions,
	fn snowdrop.TxDataFunc[T],
) (T, error) {
	return executeTx(ctx, transactionManager, opts, fn)
}
