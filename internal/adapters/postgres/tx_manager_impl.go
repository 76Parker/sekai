package postgres

import (
	"context"
	"errors"
	"sekai/pkg/ctxutils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func ExecutorFromContext(ctx context.Context, pool *pgxpool.Pool) Executor {
	if tx := ctxutils.Tx(ctx); tx != nil {
		return tx
	}
	return pool

}

type TxManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := ctxutils.Tx(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	txContext := ctxutils.WithTx(ctx, tx)
	if err := fn(txContext); err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	return tx.Commit(ctx)
}
