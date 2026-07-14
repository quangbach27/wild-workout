package domain

import (
	"errors"
	"fmt"
	"workout/common"
)

type UserType struct {
	common.Enum[UserTypeValues]
}

type UserTypeValues string

func (v UserTypeValues) Values() []string {
	return []string{"trainer", "attendee"}
}

var (
	Trainer  = common.MustEnum[UserType]("trainer")
	Attendee = common.MustEnum[UserType]("attendee")
)

type UserID string

func (id UserID) String() string {
	return string(id)
}

func (id UserID) IsZero() bool {
	return id == ""
}

type User struct {
	userID   UserID
	userType UserType
}

func (u User) ID() UserID {
	return u.userID
}

func (u User) Type() UserType {
	return u.userType
}

func (u User) IsEmpty() bool {
	return u == User{}
}

func NewUser(userID UserID, userType UserType) (User, error) {
	if userID.IsZero() {
		return User{}, errors.New("user id can't be emtpy")
	}
	if userType.IsZero() {
		return User{}, errors.New("user type can't be empty")
	}

	return User{userID: userID, userType: userType}, nil
}

func CanUserSeeTraining(user User, training *Training) error {
	if user.Type() == Trainer || user.ID() == training.UserID() {
		return nil
	}

	return fmt.Errorf("user '%s' can't see user '%s' training",
		user.userID, training.userID)
}
