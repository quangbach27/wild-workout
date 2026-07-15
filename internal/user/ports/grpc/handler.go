package grpc

import (
	"context"
	userPb "workout/common/grpc/protobuf/user"
	"workout/user/app"

	"google.golang.org/protobuf/types/known/emptypb"
)

type UserBalanceReadModel interface {
	GetUserBalance(ctx context.Context, firebaseUserID string) (int, error)
}

type Handler struct {
	userPb.UnimplementedUsersServiceServer

	appHandler           *app.Handler
	userBalanceReadModel UserBalanceReadModel
}

func NewHandler(appHandler *app.Handler, userBalanceReadModel UserBalanceReadModel) Handler {
	if appHandler == nil {
		panic("appHandler can't be nil")
	}

	if userBalanceReadModel == nil {
		panic("userBalanceReadModel can't be nil")
	}

	return Handler{
		appHandler:           appHandler,
		userBalanceReadModel: userBalanceReadModel,
	}
}

func (h Handler) GetTrainingBalance(ctx context.Context, req *userPb.GetTrainingBalanceRequest) (*userPb.GetTrainingBalanceResponse, error) {
	balance, err := h.userBalanceReadModel.GetUserBalance(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &userPb.GetTrainingBalanceResponse{
		Amount: int64(balance),
	}, nil
}

func (h Handler) UpdateTrainingBalance(ctx context.Context, req *userPb.UpdateTrainingBalanceRequest) (*emptypb.Empty, error) {
	err := h.appHandler.UpdateBalance(ctx, app.UpdateBalance{
		FirebaseUserUUID: req.UserId,
		AmountChange:     int(req.AmountChange),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
