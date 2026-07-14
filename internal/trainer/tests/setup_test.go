package tests_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
	"workout/common"
	commonHttp "workout/common/http"
	"workout/common/testutils"
	"workout/trainer"
	"workout/trainer/config"
)

var HttpBaseURL = "http://localhost:4000"

// GRPCAddr must stay in sync with the GRPC_ADDRESS set below.
var GRPCAddr = "localhost:4002"

// authToken is a stub-signed bearer token attached to every request the
// component tests send, since AuthHttpMiddleware rejects unauthenticated
// requests to any endpoint other than /health.
var authToken string

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgxDb, cleanUp := testutils.NewDB(ctx)
	defer cleanUp()

	os.Setenv("HTTP_ADDRESS", ":4000")
	// Distinct from training's own test gRPC/HTTP ports (:4001) so the two
	// packages' test binaries can run concurrently under `go test ./...`.
	os.Setenv("GRPC_ADDRESS", GRPCAddr)
	config := config.New()

	authClientStub := commonHttp.NewStubAuthClient("")
	token, err := authClientStub.NewToken(common.NewUUIDv7().String(), "Test Trainer", "trainer")
	if err != nil {
		panic(err)
	}
	authToken = token

	externalServices := trainer.ExternalServices{AuthClient: authClientStub}

	svc, err := trainer.New(ctx, pgxDb, externalServices)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := svc.Run(ctx, config.App); err != nil {
			panic(err)
		}
	}()
	waitForServer()

	os.Exit(m.Run())
}

func waitForServer() {
	for range 100 {
		resp, err := http.Get(HttpBaseURL + "/health")
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
