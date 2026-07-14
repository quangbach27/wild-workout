package grpc

import (
	"context"
	"sync"
	"workout/training/domain"
)

// StubUserGrpc is a stand-in for UserGrpc: it always succeeds and records
// the balance change it was asked to apply, so callers (e.g. component
// tests) can assert on it without needing a real user service.
type StubUserGrpc struct {
	mu             sync.Mutex
	balanceChanges map[domain.UserUUID][]int
}

func NewStubUserGrpc() *StubUserGrpc {
	return &StubUserGrpc{balanceChanges: make(map[domain.UserUUID][]int)}
}

func (s *StubUserGrpc) UpdateTrainingBalance(_ context.Context, userUUID domain.UserUUID, amountChange int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.balanceChanges[userUUID] = append(s.balanceChanges[userUUID], amountChange)
	return nil
}

// BalanceChangesFor returns every balance delta recorded for userUUID, in
// the order they were applied.
func (s *StubUserGrpc) BalanceChangesFor(userUUID domain.UserUUID) []int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]int(nil), s.balanceChanges[userUUID]...)
}
