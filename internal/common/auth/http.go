package auth

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	commonerrors "github.com/quangbach27/wild-workout/internal/common/errors"
	"github.com/quangbach27/wild-workout/internal/common/server/httperr"
)

type FirebaseHTTPMiddleware struct {
	AuthClient *auth.Client
}

func (firebaseHTTPMiddleware FirebaseHTTPMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bearerToken := firebaseHTTPMiddleware.tokenFromHeader(r)
		if bearerToken == "" {
			httperr.Unauthorised("empty-bearer-token", nil, w, r)
			return
		}

		token, err := firebaseHTTPMiddleware.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			httperr.Unauthorised("unable-to-verify-jwt", err, w, r)
			return
		}

		ctx = context.WithValue(ctx, userContextKey, User{
			UUID:        token.UID,
			Email:       token.Claims["email"].(string),
			Role:        token.Claims["role"].(string),
			DisplayName: token.Claims["name"].(string),
		})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (firebaseHTTPMiddleware FirebaseHTTPMiddleware) tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")

	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}

	return ""
}

type User struct {
	UUID  string
	Email string
	Role  string

	DisplayName string
}

type ctxKey int

const (
	userContextKey ctxKey = iota
)

var (
	NoUserInContextError = commonerrors.NewAuthorizationError("no user in context", "no-user-found")
)

func UserFromCtx(ctx context.Context) (User, error) {
	if user, ok := ctx.Value(userContextKey).(User); ok {
		return user, nil
	}

	return User{}, NoUserInContextError
}
