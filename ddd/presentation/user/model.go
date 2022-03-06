package user

import "time"

type userIDRequest struct {
	ID uint32 `uri:"id" binding:"required"`
}
type userRequest struct {
	Name string `json:"name" form:"name" binding:"required,max=32"`
}

type userResponse struct {
	ID        uint32    `json:"id"`
	Name      string    `json:"name" binding:"required,max=32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
