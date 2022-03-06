package user

type UserRepository interface {
	FindByID(id uint32) (*User, error)
	Create(user *User) (*User, error)
	Update(user *User) error
	Delete(id uint32) error
}
