package decorator

import (
	"context"
)

type CommandHandler[T any] interface {
	Execute(ctx context.Context, cmd T) error
}
