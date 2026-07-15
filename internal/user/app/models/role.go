package models

import "workout/common"

type Role struct {
	common.Enum[RoleType]
}

type RoleType string

func (rt RoleType) Values() []string { return []string{"trainer", "attendee"} }

var (
	RoleTrainer  = common.MustEnum[Role]("trainer")
	RoleAttendee = common.MustEnum[Role]("attendee")
)
