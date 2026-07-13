package tests_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
	"workout/common/testutils"
	"workout/trainer"
	"workout/trainer/config"
)

var BaseURL = "http://localhost:4000"

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgxDb, cleanUp := testutils.NewDB(ctx)
	defer cleanUp()

	os.Setenv("HTTP_ADDRESS", ":4000")
	config := config.New()

	svc, err := trainer.New(ctx, pgxDb)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := svc.Run(ctx, config); err != nil {
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
