package decorator

import "context"

// CommandHandler handles a command.
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}
