package user

import (
	"time"
)

type User struct {
	ID        uint32
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(id uint32, name string,
	createdAt time.Time, updatedAt time.Time) *User {

	return &User{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
