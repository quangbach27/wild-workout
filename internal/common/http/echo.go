package http

import (
	"log/slog"
	"net/http"
	"workout/common"

	"github.com/labstack/echo/v4"
)

func NewEcho(authClient AuthClient) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	useMiddlewares(e, authClient)

	e.HTTPErrorHandler = common.EchoErrorHandler
	e.Logger = common.NewEchoSlogAdapter(slog.Default())

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}
