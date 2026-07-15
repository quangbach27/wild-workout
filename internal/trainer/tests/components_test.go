//go:build component

package tests_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/trainer/ports/http/client"
)

var apiClient = newAPIClient()

func newAPIClient() *client.ClientWithResponses {
	c, err := client.NewClientWithResponses(HttpBaseURL, client.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+authToken)
			return nil
		},
	))
	if err != nil {
		panic(err)
	}
	return c
}

// maxValidHourDays must stay comfortably under the hour factory's 7-week
// (49-day) window. newValidHour panics if this budget is exhausted so that
// a future test which pushes the package over the limit fails loudly and
// immediately, instead of nondeterministically producing an out-of-range
// hour that the trainer service rejects with 400.
const maxValidHourDays = 45

var hourSeq atomic.Int64

// newValidHour returns a unique full-hour UTC time that satisfies the
// trainer service's hour factory constraints (see hourFactoryCfg in
// service.go): tomorrow or later, within 7 weeks, and between 08:00-17:00
// UTC. Each call is given its own day so parallel tests never see each
// other's hours when asserting on a full day's worth of results.
func newValidHour() time.Time {
	day := int(hourSeq.Add(1))
	if day > maxValidHourDays {
		panic("newValidHour: exhausted the unique-day budget within the trainer's " +
			"7-week booking window; either reuse hours across cases or raise " +
			"maxValidHourDays only after confirming it still fits the window")
	}
	hourOfDay := 8 + day%10

	base := time.Now().UTC().AddDate(0, 0, day)
	return time.Date(base.Year(), base.Month(), base.Day(), hourOfDay, 0, 0, 0, time.UTC)
}

// makeHourAvailable uses the raw (non-typed) client method rather than
// MakeHourAvailableWithResponse: the API's 204 response is declared with a
// JSON body in openapi.yaml, which is invalid for a 204 (RFC 7231) and
// results in an empty body sent with an "application/json" Content-Type.
// The typed client tries to JSON-decode that empty body and errors out, so
// tests have to go through the raw response instead.
func makeHourAvailable(t *testing.T, hours ...time.Time) {
	t.Helper()

	resp, err := apiClient.MakeHourAvailable(t.Context(), client.MakeHourAvailableJSONRequestBody{Hours: hours})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func makeHourUnavailable(t *testing.T, hours ...time.Time) {
	t.Helper()

	// See makeHourAvailable's comment: the raw client method is used here
	// too, since MakeHourUnavailableWithResponse hits the same 204-with-
	// JSON-body issue.
	resp, err := apiClient.MakeHourUnavailable(t.Context(), client.MakeHourUnavailableJSONRequestBody{Hours: hours})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func getAvailableHours(t *testing.T, from, to time.Time) []client.Date {
	t.Helper()

	resp, err := apiClient.GetTrainerAvailableHoursWithResponse(t.Context(), &client.GetTrainerAvailableHoursParams{
		DateFrom: from,
		DateTo:   to,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), string(resp.Body))
	require.NotNil(t, resp.JSON200)

	return *resp.JSON200
}

func findDate(t *testing.T, dates []client.Date, day time.Time) client.Date {
	t.Helper()

	for _, d := range dates {
		if d.Date.Year() == day.Year() && d.Date.YearDay() == day.YearDay() {
			return d
		}
	}

	t.Fatalf("date %s not found in response", day.Format("2006-01-02"))
	return client.Date{}
}

func findHour(t *testing.T, dates []client.Date, target time.Time) client.Hour {
	t.Helper()

	for _, d := range dates {
		for _, h := range d.Hours {
			if h.Hour.Equal(target) {
				return h
			}
		}
	}

	t.Fatalf("hour %s not found in response", target)
	return client.Hour{}
}

func TestMakeHourAvailable(t *testing.T) {
	t.Parallel()

	hour := newValidHour()
	makeHourAvailable(t, hour)

	dates := getAvailableHours(t, hour, hour)

	date := findDate(t, dates, hour)
	assert.True(t, date.HasFreeHours, "date should report a free hour once one is made available")

	got := findHour(t, dates, hour)
	assert.True(t, got.Available)
	assert.False(t, got.HasTrainingScheduled)
}

func TestMakeHourAvailable_IsIdempotent(t *testing.T) {
	t.Parallel()

	hour := newValidHour()
	makeHourAvailable(t, hour)
	makeHourAvailable(t, hour)

	dates := getAvailableHours(t, hour, hour)

	got := findHour(t, dates, hour)
	assert.True(t, got.Available)
}

func TestMakeHourUnavailable(t *testing.T) {
	t.Parallel()

	hour := newValidHour()
	makeHourAvailable(t, hour)

	makeHourUnavailable(t, hour)

	dates := getAvailableHours(t, hour, hour)

	got := findHour(t, dates, hour)
	assert.False(t, got.Available)
}

func TestMakeHourAvailable_RejectsInvalidHours(t *testing.T) {
	t.Parallel()

	valid := newValidHour()

	cases := []struct {
		name string
		hour time.Time
	}{
		{
			name: "in_the_past",
			hour: time.Now().UTC().AddDate(0, 0, -1),
		},
		{
			name: "not_a_full_hour",
			hour: valid.Add(30 * time.Minute),
		},
		{
			name: "before_working_hours",
			hour: time.Date(valid.Year(), valid.Month(), valid.Day(), 7, 0, 0, 0, time.UTC),
		},
		{
			name: "after_working_hours",
			hour: time.Date(valid.Year(), valid.Month(), valid.Day(), 18, 0, 0, 0, time.UTC),
		},
		{
			name: "further_than_seven_weeks_out",
			hour: time.Now().UTC().AddDate(0, 0, 50),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := apiClient.MakeHourAvailable(t.Context(), client.MakeHourAvailableJSONRequestBody{
				Hours: []time.Time{tc.hour},
			})
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		})
	}
}

func TestGetTrainerAvailableHours(t *testing.T) {
	t.Parallel()

	t.Run("dates_without_available_hours_are_still_listed", func(t *testing.T) {
		t.Parallel()

		hour := newValidHour()

		dates := getAvailableHours(t, hour, hour)

		require.Len(t, dates, 1)
		assert.False(t, dates[0].HasFreeHours)

		got := findHour(t, dates, hour)
		assert.False(t, got.Available)
	})

	t.Run("date_from_after_date_to_returns_bad_request", func(t *testing.T) {
		t.Parallel()

		hour := newValidHour()

		resp, err := apiClient.GetTrainerAvailableHoursWithResponse(t.Context(), &client.GetTrainerAvailableHoursParams{
			DateFrom: hour.AddDate(0, 0, 1),
			DateTo:   hour,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSONDefault)
		assert.Equal(t, "date-from-after-date-to", resp.JSONDefault.Slug)
	})

	t.Run("range_spans_multiple_days_and_only_marks_available_ones", func(t *testing.T) {
		t.Parallel()

		hour := newValidHour()
		makeHourAvailable(t, hour)

		from := hour.AddDate(0, 0, -1)
		to := hour.AddDate(0, 0, 1)

		dates := getAvailableHours(t, from, to)
		require.Len(t, dates, 3, "expected one entry per day in the range")

		availableDate := findDate(t, dates, hour)
		assert.True(t, availableDate.HasFreeHours)

		dayBefore := findDate(t, dates, from)
		assert.False(t, dayBefore.HasFreeHours)

		dayAfter := findDate(t, dates, to)
		assert.False(t, dayAfter.HasFreeHours)
	})
}
