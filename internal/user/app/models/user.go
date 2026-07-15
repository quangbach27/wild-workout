package models

import (
	"context"
	"workout/common"
)

type UserRepository interface {
	GetUser(ctx context.Context, uuid UserUUID) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateBalance(ctx context.Context, userID string, amountChange int) error
}

type UserUUID struct {
	common.UUID
}

// startingBalance is the training balance every new user is granted.
const startingBalance = 5

type User struct {
	uuid         UserUUID
	firebaseUUID string
	username     string
	role         Role
	balance      int
}

func (u *User) UUID() UserUUID {
	return u.uuid
}

func (u *User) FirebaseUUID() string {
	return u.firebaseUUID
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) Balance() int {
	return u.balance
}

func NewUser(firebaseUUID string, username string, role Role) (*User, error) {
	var errDetails []common.ErrorDetails

	if firebaseUUID == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "User",
			ErrorSlug:  "empty-firebase-uuid",
			Message:    "firebaseUUID cannot be empty",
		})
	}
	if username == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "User",
			ErrorSlug:  "empty-username",
			Message:    "username cannot be empty",
		})
	}
	if role.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "User",
			ErrorSlug:  "empty-role",
			Message:    "role cannot be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError(
			"invalid-user",
			"user is not valid",
		).WithDetails(errDetails)
	}

	return &User{
		uuid:         UserUUID{common.NewUUIDv7()},
		firebaseUUID: firebaseUUID,
		username:     username,
		role:         role,
		balance:      startingBalance,
	}, nil
}

// UnmarshalUserFromDB rebuilds a User from data already persisted in the
// database. It skips the validation NewUser performs because data coming
// from our own storage is trusted.
func UnmarshalUserFromDB(uuid UserUUID, firebaseUUID string, username string, role Role, balance int) *User {
	return &User{
		uuid:         uuid,
		firebaseUUID: firebaseUUID,
		username:     username,
		role:         role,
		balance:      balance,
	}
}

func (u *User) UpdateBalance(amountChange int) error {
	newBalance := u.balance + amountChange
	if newBalance < 0 {
		return common.NewInvalidInputError("insufficient-balance", "balance can't be smaller than 0")
	}

	u.balance = newBalance
	return nil
}
