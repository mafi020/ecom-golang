package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

// Transactor handles the transaction lifecycle
type Transactor struct {
	db *sql.DB
}

func NewPostgresTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	// Inject tx into context
	txCtx := context.WithValue(ctx, txKey{}, tx)

	err = fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

// GetExecutor extracts the tx from context or returns the standard db pool
func GetExecutor(ctx context.Context, db *sql.DB) DBExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

// DBExecutor is an interface that both *sql.DB and *sql.Tx satisfy
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
