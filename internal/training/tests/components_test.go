package tests_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/training/domain"
	"workout/training/ports/http/client"
)

func TestCreateTraining(t *testing.T) {
	t.Parallel()

	attendee, attendeeID := newAuthenticatedClient(t, "attendee")
	trainingTime := newTrainingTime()

	createTraining(t, attendee, trainingTime, "leg day")

	trainings := getTrainings(t, attendee)
	got := findTrainingByTime(t, trainings, trainingTime)

	assert.Equal(t, "leg day", got.Notes)
	assert.Equal(t, attendeeID, got.UserId)
	assert.True(t, got.CanBeCancelled)
	assert.False(t, got.MoveRequiresAccept)

	assert.Equal(t, []int{-1}, userService.BalanceChangesFor(attendeeID))
	assert.True(t, trainerService.WasScheduled(trainingTime))
}

func TestCreateTraining_RejectsNoteTooLong(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")

	resp, err := attendee.CreateTrainingWithResponse(t.Context(), client.PostTraining{
		Time:  newTrainingTime(),
		Notes: string(make([]byte, 1001)),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
	require.NotNil(t, resp.JSONDefault)
	assert.Equal(t, "note-too-long", resp.JSONDefault.Slug)
}

func TestCreateTraining_RequiresAuthentication(t *testing.T) {
	t.Parallel()

	resp, err := http.Get(BaseURL + "/trainings")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetTrainings_AttendeeOnlySeesOwnTrainings(t *testing.T) {
	t.Parallel()

	alice, aliceID := newAuthenticatedClient(t, "attendee")
	bob, _ := newAuthenticatedClient(t, "attendee")

	aliceTraining := newTrainingTime()
	bobTraining := newTrainingTime()

	createTraining(t, alice, aliceTraining, "alice's training")
	createTraining(t, bob, bobTraining, "bob's training")

	aliceTrainings := getTrainings(t, alice)
	for _, tr := range aliceTrainings {
		assert.Equal(t, aliceID, tr.UserId, "attendee should only see their own trainings")
	}
	findTrainingByTime(t, aliceTrainings, aliceTraining)
}

func TestGetTrainings_TrainerSeesAllTrainings(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")
	trainer, _ := newAuthenticatedClient(t, "trainer")

	trainingTime := newTrainingTime()
	createTraining(t, attendee, trainingTime, "trainer visibility check")

	trainerView := getTrainings(t, trainer)
	findTrainingByTime(t, trainerView, trainingTime)
}

func TestCancelTraining(t *testing.T) {
	t.Parallel()

	attendee, attendeeID := newAuthenticatedClient(t, "attendee")
	trainingTime := newTrainingTime()
	createTraining(t, attendee, trainingTime, "cancel me")

	created := findTrainingByTime(t, getTrainings(t, attendee), trainingTime)

	resp, err := attendee.CancelTrainingWithResponse(t.Context(), created.Uuid)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), string(resp.Body))

	for _, tr := range getTrainings(t, attendee) {
		assert.NotEqual(t, created.Uuid, tr.Uuid, "cancelled training should no longer be listed")
	}

	assert.Equal(t, []int{-1, 1}, userService.BalanceChangesFor(attendeeID), "free cancellation should refund the training")
	assert.True(t, trainerService.WasCancelled(trainingTime))
}

func TestCancelTraining_ForbiddenForOtherAttendee(t *testing.T) {
	t.Parallel()

	owner, _ := newAuthenticatedClient(t, "attendee")
	other, _ := newAuthenticatedClient(t, "attendee")

	trainingTime := newTrainingTime()
	createTraining(t, owner, trainingTime, "owner only")
	created := findTrainingByTime(t, getTrainings(t, owner), trainingTime)

	resp, err := other.CancelTrainingWithResponse(t.Context(), created.Uuid)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode())
	require.NotNil(t, resp.JSONDefault)
	assert.Equal(t, "forbidden", resp.JSONDefault.Slug)
}

func TestRescheduleTraining(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")
	originalTime := newTrainingTime()
	newTime := newTrainingTime()

	createTraining(t, attendee, originalTime, "reschedule me")
	created := findTrainingByTime(t, getTrainings(t, attendee), originalTime)

	resp, err := attendee.RescheduleTrainingWithResponse(t.Context(), created.Uuid, client.PostTraining{
		Time:  newTime,
		Notes: "rescheduled",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), string(resp.Body))

	got := findTrainingByTime(t, getTrainings(t, attendee), newTime)
	assert.Equal(t, "rescheduled", got.Notes)
	assert.True(t, trainerService.WasMoved(newTime, originalTime))
}

func TestRescheduleTraining_RejectedWhenTooCloseToStart(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")
	soonTime := newSoonTrainingTime()

	createTraining(t, attendee, soonTime, "too late to move")
	created := findTrainingByTime(t, getTrainings(t, attendee), soonTime)

	resp, err := attendee.RescheduleTrainingWithResponse(t.Context(), created.Uuid, client.PostTraining{
		Time:  newTrainingTime(),
		Notes: "too late",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
}

func TestRequestAndApproveRescheduleTraining(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")
	trainer, _ := newAuthenticatedClient(t, "trainer")

	originalTime := newTrainingTime()
	proposedTime := newTrainingTime()

	createTraining(t, attendee, originalTime, "propose reschedule")
	created := findTrainingByTime(t, getTrainings(t, attendee), originalTime)

	reqResp, err := attendee.RequestRescheduleTrainingWithResponse(t.Context(), created.Uuid, client.PostTraining{
		Time:  proposedTime,
		Notes: created.Notes,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, reqResp.StatusCode(), string(reqResp.Body))

	pending := findTrainingByTime(t, getTrainings(t, attendee), originalTime)
	require.NotNil(t, pending.ProposedTime)
	assert.True(t, pending.ProposedTime.Equal(proposedTime))
	assert.True(t, pending.MoveRequiresAccept)
	require.NotNil(t, pending.MoveProposedBy)
	assert.Equal(t, domain.Attendee, *pending.MoveProposedBy)

	approveResp, err := trainer.ApproveRescheduleTrainingWithResponse(t.Context(), created.Uuid)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, approveResp.StatusCode(), string(approveResp.Body))

	approved := findTrainingByTime(t, getTrainings(t, attendee), proposedTime)
	assert.False(t, approved.MoveRequiresAccept)
	assert.Nil(t, approved.MoveProposedBy)
	assert.True(t, trainerService.WasMoved(proposedTime, originalTime))
}

func TestRequestAndRejectRescheduleTraining(t *testing.T) {
	t.Parallel()

	attendee, _ := newAuthenticatedClient(t, "attendee")
	trainer, _ := newAuthenticatedClient(t, "trainer")

	originalTime := newTrainingTime()
	proposedTime := newTrainingTime()

	createTraining(t, attendee, originalTime, "propose then reject")
	created := findTrainingByTime(t, getTrainings(t, attendee), originalTime)

	reqResp, err := attendee.RequestRescheduleTrainingWithResponse(t.Context(), created.Uuid, client.PostTraining{
		Time:  proposedTime,
		Notes: created.Notes,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, reqResp.StatusCode(), string(reqResp.Body))

	rejectResp, err := trainer.RejectRescheduleTrainingWithResponse(t.Context(), created.Uuid)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rejectResp.StatusCode(), string(rejectResp.Body))

	got := findTrainingByTime(t, getTrainings(t, attendee), originalTime)
	assert.False(t, got.MoveRequiresAccept)
	assert.Nil(t, got.MoveProposedBy)
	assert.Nil(t, got.ProposedTime)
}
