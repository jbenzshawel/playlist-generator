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

type dbTransactionDecorator[TParam, TRes any] struct {
	base CommandWithResultHandler[TParam, TRes]
	tx   transaction
}

func (d dbTransactionDecorator[TParam, TRes]) Execute(ctx context.Context, cmd TParam) (TRes, error) {
	var zero TRes
	err := d.tx.Begin(ctx)
	if err != nil {
		return zero, err
	}

	res, err := d.base.Execute(ctx, cmd)
	if err != nil {
		txErr := d.tx.Rollback()
		if txErr != nil {
			slog.Warn("transaction rollback failed", slog.Any("error", txErr))
		}
		return zero, err
	}

	err = d.tx.Commit()
	if err != nil {
		return zero, err
	}
	return res, nil
}

func ApplyDBTransactionDecorator[TParam, TRes any](c CommandWithResultHandler[TParam, TRes], tx transaction) CommandWithResultHandler[TParam, TRes] {
	return &dbTransactionDecorator[TParam, TRes]{
		base: c,
		tx:   tx,
	}
}
