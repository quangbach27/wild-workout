package http

import (
	"context"
	"workout/common"
	"workout/user/app"
)

type Handler struct {
	appHandler *app.Handler
}

func NewHandler(appHandler *app.Handler) Handler {
	if appHandler == nil {
		panic("appHandler can't be nil")
	}

	return Handler{
		appHandler: appHandler,
	}
}

func Register(router common.EchoRouter, handler Handler) error {
	RegisterHandlers(router, NewStrictHandler(handler, nil))
	return nil
}

func (h Handler) GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error) {
	return GetCurrentUser200JSONResponse{
		Balance:     1,
		DisplayName: "",
		Role:        "",
	}, nil
}
