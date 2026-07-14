package http

import (
	"context"
	"errors"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"

	"github.com/golang-jwt/jwt/v5"
)

const defaultStubJWTSecret = "wild-workouts-stub-jwt-secret-do-not-use-in-production"

// StubAuthClient mints and verifies its own HMAC-signed tokens instead of
// calling Firebase, so local development and tests can authenticate without
// a real Firebase project. It satisfies AuthClient, so wiring it into
// FirebaseHttpMiddleware and later swapping in a real *auth.Client is a
// one-line change.
type StubAuthClient struct {
	secret []byte
}

// NewStubAuthClient builds a StubAuthClient. secret is the HMAC key tokens
// are signed and verified with; if empty, it falls back to the
// JWT_STUB_SECRET env var, then to a hardcoded dev default.
func NewStubAuthClient(secret string) *StubAuthClient {
	if secret == "" {
		secret = os.Getenv("JWT_STUB_SECRET")
	}
	if secret == "" {
		secret = defaultStubJWTSecret
	}

	return &StubAuthClient{secret: []byte(secret)}
}

type stubClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// NewToken mints a token this client's VerifyIDToken will accept, encoding
// userID as the Firebase uid/sub and name/role as custom claims.
func (c *StubAuthClient) NewToken(userID, name, role string) (string, error) {
	now := time.Now()
	claims := stubClaims{
		Name: name,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(c.secret)
}

func (c *StubAuthClient) VerifyIDToken(_ context.Context, idToken string) (*auth.Token, error) {
	var claims stubClaims

	_, err := jwt.ParseWithClaims(idToken, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return c.secret, nil
	})
	if err != nil {
		return nil, err
	}

	return &auth.Token{
		UID:     claims.Subject,
		Subject: claims.Subject,
		Claims: map[string]interface{}{
			"name": claims.Name,
			"role": claims.Role,
		},
	}, nil
}
