package tests_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
	"workout/common"
	"workout/training/domain"
	"workout/training/ports/http/client"

	"github.com/stretchr/testify/require"
)

// newAuthenticatedClient mints a fresh stub-signed bearer token for a new
// user of the given role and returns a client that attaches it to every
// request, along with the UUID the token authenticates as.
func newAuthenticatedClient(t *testing.T, role string) (*client.ClientWithResponses, domain.UserUUID) {
	t.Helper()

	userUUID := domain.UserUUID{UUID: common.NewUUIDv7()}

	token, err := stubAuthClient.NewToken(userUUID.String(), "Test "+role, role)
	require.NoError(t, err)

	c, err := client.NewClientWithResponses(BaseURL, client.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		},
	))
	require.NoError(t, err)

	return c, userUUID
}

var trainingDaySeq atomic.Int64

// newTrainingTime returns a unique training time, far enough in the future
// (>24h) that free-cancellation/reschedule rules don't kick in unless a
// test deliberately wants to exercise them. Each call gets its own day so
// parallel tests never collide.
func newTrainingTime() time.Time {
	day := trainingDaySeq.Add(1)
	return time.Now().UTC().AddDate(0, 0, int(day)+2).Truncate(time.Second)
}

var soonTrainingSeq atomic.Int64

// newSoonTrainingTime returns a unique training time within the next 24h,
// for tests exercising the "too late to change for free" rules.
func newSoonTrainingTime() time.Time {
	n := soonTrainingSeq.Add(1)
	return time.Now().UTC().Add(time.Duration(n) * time.Minute).Truncate(time.Second)
}

func createTraining(t *testing.T, c *client.ClientWithResponses, trainingTime time.Time, notes string) {
	t.Helper()

	resp, err := c.CreateTrainingWithResponse(t.Context(), client.PostTraining{
		Time:  trainingTime,
		Notes: notes,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), string(resp.Body))
}

func getTrainings(t *testing.T, c *client.ClientWithResponses) []client.Training {
	t.Helper()

	resp, err := c.GetTrainingsWithResponse(t.Context())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON200)

	return resp.JSON200.Trainings
}

func findTrainingByTime(t *testing.T, trainings []client.Training, trainingTime time.Time) client.Training {
	t.Helper()

	for _, tr := range trainings {
		if tr.Time.Equal(trainingTime) {
			return tr
		}
	}

	t.Fatalf("training at %s not found", trainingTime)
	return client.Training{}
}
