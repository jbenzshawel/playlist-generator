package decorator

import (
	"context"
	"log/slog"
)

type transaction interface {
	Begin(ctx context.Context) error
	Rollback() error
	Commit() error
}

type dbTransactionDecorator[T any] struct {
	base CommandHandler[T]
	tx   transaction
}

func (d dbTransactionDecorator[T]) Execute(ctx context.Context, cmd T) error {
	err := d.tx.Begin(ctx)
	if err != nil {
		return err
	}

	err = d.base.Execute(ctx, cmd)
	if err != nil {
		txErr := d.tx.Rollback()
		if txErr != nil {
			slog.Warn("transaction rollback failed", slog.Any("error", txErr))
		}
		return err
	}

	return d.tx.Commit()
}

func ApplyDBTransactionDecorator[T any](c CommandHandler[T], tx transaction) CommandHandler[T] {
	return &dbTransactionDecorator[T]{
		base: c,
		tx:   tx,
	}
}
