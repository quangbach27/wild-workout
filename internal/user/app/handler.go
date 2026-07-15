package app

import "workout/user/app/models"

type Handler struct {
	userRepo models.UserRepository
}

func NewHandler(userRepo models.UserRepository) *Handler {
	if userRepo == nil {
		panic("userRepo can't be empty")
	}

	return &Handler{
		userRepo: userRepo,
	}
}
