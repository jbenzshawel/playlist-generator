package decorator

import (
	"context"
)

type CommandWithResultHandler[TParam, TRes any] interface {
	Execute(ctx context.Context, cmd TParam) (TRes, error)
}

type CommandHandler[TParam any] interface {
	CommandWithResultHandler[TParam, any]
}
