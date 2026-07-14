package http

import (
	"context"
	"strings"
	"workout/common"

	"firebase.google.com/go/v4/auth"

	"github.com/labstack/echo/v4"
)

// AuthClient verifies a Firebase ID token and returns the identity it
// encodes. *firebase.google.com/go/v4/auth.Client satisfies this directly;
// StubAuthClient is a drop-in replacement for local development and tests.
type AuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// AuthClaims is the identity extracted from a request's bearer token.
type AuthClaims struct {
	UserID string
	Name   string
	Role   string
}

type authContextKey struct{}

func ContextWithAuthClaims(ctx context.Context, claims AuthClaims) context.Context {
	return context.WithValue(ctx, authContextKey{}, claims)
}

func AuthClaimsFromContext(ctx context.Context) (AuthClaims, bool) {
	claims, ok := ctx.Value(authContextKey{}).(AuthClaims)
	return claims, ok
}

// AuthHttpMiddleware verifies the request's bearer ID token via
// AuthClient and puts the resulting AuthClaims on the request context.
type AuthHttpMiddleware struct {
	AuthClient AuthClient
}

func NewAuthHttpMiddleware(client AuthClient) AuthHttpMiddleware {
	if client == nil {
		panic("AuthClient can't be nil")
	}

	return AuthHttpMiddleware{AuthClient: client}
}

func (m AuthHttpMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().URL.Path == "/health" {
			return next(c)
		}

		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		idToken, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || idToken == "" {
			return common.NewUnauthorizedError("invalid-token", "missing bearer token")
		}

		token, err := m.AuthClient.VerifyIDToken(c.Request().Context(), idToken)
		if err != nil {
			return common.NewUnauthorizedError("invalid-token", "verify id token failed")
		}

		claims := AuthClaims{UserID: token.UID}
		if name, ok := token.Claims["name"].(string); ok {
			claims.Name = name
		}
		if role, ok := token.Claims["role"].(string); ok {
			claims.Role = role
		}

		ctx := ContextWithAuthClaims(c.Request().Context(), claims)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
