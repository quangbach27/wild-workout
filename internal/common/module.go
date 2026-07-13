package common

import (
	"context"
)

type Module interface {
	Name() string
	Init(ctx context.Context) error
	RegisterHttp(ctx context.Context, e EchoRouter) error
	RegisterGrpc(ctx context.Context) error
}
