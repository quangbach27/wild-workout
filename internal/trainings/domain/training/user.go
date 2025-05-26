package training

import (
	"fmt"

	commonErrors "github.com/quangbach27/wild-workout/internal/common/errors"
)

type UserType struct {
	s string
}

func (userType UserType) IsZero() bool {
	return userType == UserType{}
}

func (userType UserType) String() string {
	return userType.s
}

var (
	Trainer  = UserType{"trainer"}
	Attendee = UserType{"attendee"}
)

var userTypes = []UserType{
	Trainer,
	Attendee,
}

func NewUserTypeFromString(userType string) (UserType, error) {
	for _, u := range userTypes {
		if u.s == userType {
			return u, nil
		}
	}
	return UserType{}, commonErrors.NewSlugError(
		fmt.Sprintf("invalid '%s' role", userType),
		"invalid-role",
	)
}
