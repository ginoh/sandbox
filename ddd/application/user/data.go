package user

import (
	"time"

	"example.com/domain/model/user"
)

type UserData struct {
	ID        uint32
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserData(user *user.User) *UserData {
	return &UserData{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
