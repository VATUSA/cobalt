package dbconn

import (
	"context"
	"vatusa-cobalt/db"
)

func WithTransaction(ctx context.Context, fn func(*db.Queries) error) error {
	_, err := WithTransactionResult(ctx, func(q *db.Queries) (struct{}, error) {
		return struct{}{}, fn(q)
	})
	return err
}

func WithTransactionResult[T any](ctx context.Context, fn func(*db.Queries) (T, error)) (T, error) {
	var zero T
	tx, err := DB().BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback() // no-op after successful Commit
	result, err := fn(db.New(tx))
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	return result, nil
}
