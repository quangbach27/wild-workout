package http

import (
	"context"
	"workout/common"
	commonHttp "workout/common/http"
	"workout/training/domain"
)

// authenticatedUser is the caller identity resolved from the request's auth
// claims: domain.User for authorization checks, Name for anything that needs
// a display name (e.g. creating a training).
type authenticatedUser struct {
	domain.User
	Name string
}

func userFromContext(ctx context.Context) (authenticatedUser, error) {
	claims, ok := commonHttp.AuthClaimsFromContext(ctx)
	if !ok {
		return authenticatedUser{}, common.NewUnauthorizedError("unauthorized", "missing authentication")
	}

	var userType domain.UserType
	if err := userType.UnmarshalText([]byte(claims.Role)); err != nil {
		return authenticatedUser{}, common.NewUnauthorizedError("unauthorized", "invalid role in token")
	}

	user, err := domain.NewUser(domain.UserID(claims.UserID), userType)
	if err != nil {
		return authenticatedUser{}, common.NewUnauthorizedError("unauthorized", "%s", err.Error())
	}

	return authenticatedUser{User: user, Name: claims.Name}, nil
}
