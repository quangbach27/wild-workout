package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/common"
	"workout/training/domain"
)

func testUserID(t *testing.T) domain.UserID {
	t.Helper()

	return domain.UserID(common.NewUUIDv7().String())
}

func newTraining(t *testing.T, trainingTime time.Time) *domain.Training {
	t.Helper()

	tr, err := domain.NewTraining(testUserID(t), "user-name", trainingTime)
	require.NoError(t, err)

	return tr
}

func TestNewTraining(t *testing.T) {
	t.Parallel()

	trainingTime := time.Now().Add(48 * time.Hour)
	userID := testUserID(t)

	testCases := []struct {
		Name         string
		UserID       domain.UserID
		UserName     string
		TrainingTime time.Time
		ExpectError  bool
	}{
		{
			Name:         "valid",
			UserID:       userID,
			UserName:     "user-name",
			TrainingTime: trainingTime,
		},
		{
			Name:         "empty_user_id",
			UserID:       domain.UserID(""),
			UserName:     "user-name",
			TrainingTime: trainingTime,
			ExpectError:  true,
		},
		{
			Name:         "empty_user_name",
			UserID:       userID,
			UserName:     "",
			TrainingTime: trainingTime,
			ExpectError:  true,
		},
		{
			Name:         "zero_training_time",
			UserID:       userID,
			UserName:     "user-name",
			TrainingTime: time.Time{},
			ExpectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr, err := domain.NewTraining(tc.UserID, tc.UserName, tc.TrainingTime)

			if tc.ExpectError {
				require.Error(t, err)
				assert.Nil(t, tr)

				var commonErr common.Error
				require.ErrorAs(t, err, &commonErr)
				assert.NotEmpty(t, commonErr.Details)
				return
			}

			require.NoError(t, err)
			assert.False(t, tr.UUID().IsZero())
			assert.Equal(t, tc.UserID, tr.UserID())
			assert.Equal(t, tc.UserName, tr.UserName())
			assert.True(t, tc.TrainingTime.Equal(tr.Time()))
			assert.False(t, tr.IsCanceled())
			assert.False(t, tr.IsRescheduleProposed())
		})
	}
}

func TestTraining_CanBeCanceledForFree(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Time     time.Time
		Expected bool
	}{
		{
			Name:     "far_in_the_future",
			Time:     time.Now().Add(48 * time.Hour),
			Expected: true,
		},
		{
			Name:     "less_than_24h",
			Time:     time.Now().Add(1 * time.Hour),
			Expected: false,
		},
		{
			Name:     "in_the_past",
			Time:     time.Now().Add(-1 * time.Hour),
			Expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, tc.Time)

			assert.Equal(t, tc.Expected, tr.CanBeCanceledForFree())
		})
	}
}

func TestTraining_ProposeReschedule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name         string
		ProposedBy   domain.UserType
		ProposedTime time.Time
	}{
		{
			Name:         "proposed_by_attendee",
			ProposedBy:   domain.Attendee,
			ProposedTime: time.Now().Add(72 * time.Hour),
		},
		{
			Name:         "proposed_by_trainer",
			ProposedBy:   domain.Trainer,
			ProposedTime: time.Now().Add(96 * time.Hour),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, time.Now().Add(48*time.Hour))

			tr.ProposeReschedule(tc.ProposedTime, tc.ProposedBy)

			assert.True(t, tr.IsRescheduleProposed())
			assert.True(t, tc.ProposedTime.Equal(tr.ProposedNewTime()))
			assert.Equal(t, tc.ProposedBy, tr.MoveProposedBy())
		})
	}
}

func TestTraining_RescheduleTraining(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name        string
		TrainingIn  time.Duration
		ExpectError bool
	}{
		{
			Name:       "success",
			TrainingIn: 48 * time.Hour,
		},
		{
			Name:        "error_not_enough_time",
			TrainingIn:  1 * time.Hour,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, time.Now().Add(tc.TrainingIn))
			originalTime := tr.Time()
			newTime := time.Now().Add(72 * time.Hour)

			err := tr.RescheduleTraining(newTime)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.True(t, originalTime.Equal(tr.Time()))
				return
			}

			require.NoError(t, err)
			assert.True(t, newTime.Equal(tr.Time()))
		})
	}
}

func TestTraining_ApproveReschedule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name        string
		Propose     bool
		ProposedBy  domain.UserType
		ApproveBy   domain.UserType
		ExpectError bool
	}{
		{
			Name:       "success",
			Propose:    true,
			ProposedBy: domain.Attendee,
			ApproveBy:  domain.Trainer,
		},
		{
			Name:        "error_not_proposed",
			Propose:     false,
			ApproveBy:   domain.Trainer,
			ExpectError: true,
		},
		{
			Name:        "error_same_user_type",
			Propose:     true,
			ProposedBy:  domain.Attendee,
			ApproveBy:   domain.Attendee,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, time.Now().Add(48*time.Hour))
			newTime := time.Now().Add(72 * time.Hour)

			if tc.Propose {
				tr.ProposeReschedule(newTime, tc.ProposedBy)
			}

			err := tr.ApproveReschedule(tc.ApproveBy)

			if tc.ExpectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, newTime.Equal(tr.Time()))
			assert.False(t, tr.IsRescheduleProposed())
			assert.True(t, tr.ProposedNewTime().IsZero())
			assert.True(t, tr.MoveProposedBy().IsZero())
		})
	}
}

func TestTraining_RejectReschedule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name        string
		Propose     bool
		ExpectError bool
	}{
		{
			Name:    "success",
			Propose: true,
		},
		{
			Name:        "error_not_proposed",
			Propose:     false,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, time.Now().Add(48*time.Hour))

			if tc.Propose {
				tr.ProposeReschedule(time.Now().Add(72*time.Hour), domain.Attendee)
			}

			err := tr.RejectReschedule()

			if tc.ExpectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.False(t, tr.IsRescheduleProposed())
			assert.True(t, tr.ProposedNewTime().IsZero())
			assert.True(t, tr.MoveProposedBy().IsZero())
		})
	}
}

func TestTraining_Cancel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name            string
		AlreadyCanceled bool
		ExpectError     bool
	}{
		{
			Name:            "success",
			AlreadyCanceled: false,
		},
		{
			Name:            "error_already_canceled",
			AlreadyCanceled: true,
			ExpectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			tr := newTraining(t, time.Now().Add(48*time.Hour))
			if tc.AlreadyCanceled {
				require.NoError(t, tr.Cancel())
			}

			err := tr.Cancel()

			if tc.ExpectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, tr.IsCanceled())
		})
	}
}
