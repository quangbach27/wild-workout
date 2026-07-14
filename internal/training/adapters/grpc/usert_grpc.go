package grpc

import (
	"context"
	"workout/common/grpc/protobuf/user"
	"workout/training/domain"
)

type UserGrpc struct {
	client user.UsersServiceClient
}

func NewUsersGrpc(client user.UsersServiceClient) *UserGrpc {
	if client == nil {
		panic("userServiceClient can't be nil")
	}
	return &UserGrpc{client: client}
}

func (s *UserGrpc) UpdateTrainingBalance(ctx context.Context, userID domain.UserID, amountChange int) error {
	_, err := s.client.UpdateTrainingBalance(ctx, &user.UpdateTrainingBalanceRequest{
		UserId:       userID.String(),
		AmountChange: int64(amountChange),
	})

	return err
}
