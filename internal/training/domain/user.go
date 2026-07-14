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

type UserUUID struct {
	common.UUID
}

type User struct {
	userUUID UserUUID
	userType UserType
}

func (u User) UUID() UserUUID {
	return u.userUUID
}

func (u User) Type() UserType {
	return u.userType
}

func (u User) IsEmpty() bool {
	return u == User{}
}

func NewUser(userUUID UserUUID, userType UserType) (User, error) {
	if userUUID.IsZero() {
		return User{}, errors.New("user uuid can't be emtpy")
	}
	if userType.IsZero() {
		return User{}, errors.New("user type can't be empty")
	}

	return User{userUUID: userUUID, userType: userType}, nil
}

func CanUserSeeTraining(user User, training *Training) error {
	if user.Type() == Trainer || user.UUID() == training.UserUUID() {
		return nil
	}

	return fmt.Errorf("user '%s' can't see user '%s' training",
		user.userUUID, training.userUUID)
}
