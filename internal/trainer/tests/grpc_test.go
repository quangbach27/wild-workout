//go:build component

package tests_test

import (
	"testing"
	commonGrpc "workout/common/grpc"
	trainerpb "workout/common/grpc/protobuf/trainer"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcClient = newGRPCClient()

func newGRPCClient() trainerpb.TrainerServiceClient {
	conn, err := commonGrpc.NewGRPCClientConn(GRPCAddr)
	if err != nil {
		panic(err)
	}
	return trainerpb.NewTrainerServiceClient(conn)
}

func isHourAvailable(t *testing.T, hourTime *timestamppb.Timestamp) bool {
	t.Helper()

	resp, err := grpcClient.IsHourAvailable(t.Context(), &trainerpb.IsHourAvailableRequest{Time: hourTime})
	require.NoError(t, err)

	return resp.GetIsAvailable()
}

func TestGRPC_IsHourAvailable_ReflectsUnavailableHour(t *testing.T) {
	t.Parallel()

	hour := newValidHour()
	// The local dev DB persists across test runs, so explicitly put the
	// hour in a known "not available" state rather than assuming it was
	// never touched by an earlier run.
	makeHourUnavailable(t, hour)

	require.False(t, isHourAvailable(t, timestamppb.New(hour)))
}

func TestGRPC_MakeHourAvailable(t *testing.T) {
	t.Parallel()

	hourTime := timestamppb.New(newValidHour())

	_, err := grpcClient.MakeHourAvailable(t.Context(), &trainerpb.UpdateHourRequest{Time: hourTime})
	require.NoError(t, err)

	require.True(t, isHourAvailable(t, hourTime))
}

// TestGRPC_ScheduleAndCancelTraining exercises the trainer gRPC server end
// to end: make an hour available, schedule a training on it over gRPC, then
// cancel it, checking IsHourAvailable after each step.
func TestGRPC_ScheduleAndCancelTraining(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	hourTime := timestamppb.New(newValidHour())

	_, err := grpcClient.MakeHourAvailable(ctx, &trainerpb.UpdateHourRequest{Time: hourTime})
	require.NoError(t, err)
	require.True(t, isHourAvailable(t, hourTime))

	_, err = grpcClient.ScheduleTraining(ctx, &trainerpb.UpdateHourRequest{Time: hourTime})
	require.NoError(t, err)
	require.False(t, isHourAvailable(t, hourTime), "hour should no longer be available once a training is scheduled")

	_, err = grpcClient.CancelTraining(ctx, &trainerpb.UpdateHourRequest{Time: hourTime})
	require.NoError(t, err)
	require.True(t, isHourAvailable(t, hourTime), "hour should be available again once the training is cancelled")
}

func TestGRPC_ScheduleTraining_RejectsHourThatIsNotAvailable(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	hour := newValidHour()
	makeHourUnavailable(t, hour)

	_, err := grpcClient.ScheduleTraining(ctx, &trainerpb.UpdateHourRequest{Time: timestamppb.New(hour)})
	require.Error(t, err)
}

func TestGRPC_CancelTraining_RejectsHourWithNoScheduledTraining(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	hourTime := timestamppb.New(newValidHour())

	_, err := grpcClient.MakeHourAvailable(ctx, &trainerpb.UpdateHourRequest{Time: hourTime})
	require.NoError(t, err)

	// The hour is available but nothing was ever scheduled on it.
	_, err = grpcClient.CancelTraining(ctx, &trainerpb.UpdateHourRequest{Time: hourTime})
	require.Error(t, err)
}
