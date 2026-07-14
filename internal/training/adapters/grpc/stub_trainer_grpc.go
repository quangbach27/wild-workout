package grpc

import (
	"context"
	"sync"
	"time"
)

// StubTrainerGrpc is a stand-in for TrainerGrpc: it always succeeds and
// records every call, so callers (e.g. component tests) can assert the
// trainer's calendar would have been kept in sync without needing a real
// trainer service.
type StubTrainerGrpc struct {
	mu        sync.Mutex
	scheduled []time.Time
	cancelled []time.Time
	moved     []MovedTraining
}

// MovedTraining is one recorded call to StubTrainerGrpc.MoveTraining.
type MovedTraining struct {
	NewTime      time.Time
	OriginalTime time.Time
}

func NewStubTrainerGrpc() *StubTrainerGrpc {
	return &StubTrainerGrpc{}
}

func (s *StubTrainerGrpc) ScheduleTraining(_ context.Context, trainingTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.scheduled = append(s.scheduled, trainingTime)
	return nil
}

func (s *StubTrainerGrpc) CancelTraining(_ context.Context, trainingTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancelled = append(s.cancelled, trainingTime)
	return nil
}

func (s *StubTrainerGrpc) MoveTraining(_ context.Context, newTime, originalTrainingTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.moved = append(s.moved, MovedTraining{NewTime: newTime, OriginalTime: originalTrainingTime})
	return nil
}

// WasScheduled reports whether ScheduleTraining was called with trainingTime.
func (s *StubTrainerGrpc) WasScheduled(trainingTime time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.scheduled {
		if t.Equal(trainingTime) {
			return true
		}
	}
	return false
}

// WasCancelled reports whether CancelTraining was called with trainingTime.
func (s *StubTrainerGrpc) WasCancelled(trainingTime time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.cancelled {
		if t.Equal(trainingTime) {
			return true
		}
	}
	return false
}

// WasMoved reports whether MoveTraining was called with this newTime/
// originalTime pair.
func (s *StubTrainerGrpc) WasMoved(newTime, originalTime time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range s.moved {
		if m.NewTime.Equal(newTime) && m.OriginalTime.Equal(originalTime) {
			return true
		}
	}
	return false
}
