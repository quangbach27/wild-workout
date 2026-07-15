//go:build component

package tests_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
	commonHttp "workout/common/http"
	"workout/common/testutils"
	"workout/training"
	grpcAdapter "workout/training/adapters/grpc"
	"workout/training/config"
)

var BaseURL = "http://localhost:4001"

// stubAuthClient signs the bearer tokens component tests authenticate with;
// see newAuthenticatedClient in helpers_test.go.
var stubAuthClient *commonHttp.StubAuthClient

// userService and trainerService are shared across every test in this
// package so tests can assert on the calls the training service made to its
// external dependencies (e.g. balance changes, scheduled/cancelled times).
var (
	userService    = grpcAdapter.NewStubUserGrpc()
	trainerService = grpcAdapter.NewStubTrainerGrpc()
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgxDb, cleanUp := testutils.NewDB(ctx)
	defer cleanUp()

	os.Setenv("HTTP_ADDRESS", ":4001")
	cfg := config.New()

	stubAuthClient = commonHttp.NewStubAuthClient("")

	externalServices := training.ExternalServices{
		UserService:    userService,
		TrainerService: trainerService,
		AuthClient:     stubAuthClient,
	}

	svc, err := training.NewService(ctx, pgxDb, externalServices)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := svc.Run(ctx, cfg.App); err != nil {
			panic(err)
		}
	}()
	waitForServer()

	os.Exit(m.Run())
}

func waitForServer() {
	for range 100 {
		resp, err := http.Get(BaseURL + "/health")
		if err == nil && resp.StatusCode < 300 {
			_ = resp.Body.Close()
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	panic("server did not start within 5 seconds")
}
