package user

import (
	"example.com/domain/model/user"
)

type UserService interface {
	FindByID(id uint32) (*UserData, error)
	Create(user *UserData) (*UserData, error)
	Update(user *UserData) error
	Delete(id uint32) error
}

type userService struct {
	user.UserRepository
}

func NewUserService(repo user.UserRepository) UserService {
	return &userService{repo}
}

func (us *userService) FindByID(id uint32) (*UserData, error) {
	u, err := us.UserRepository.FindByID(id)
	if err != nil {
		return nil, err
	}
	return NewUserData(u), nil
}

func (us *userService) Create(ud *UserData) (*UserData, error) {
	u, err := us.UserRepository.Create(newUserFromData(ud))
	if err != nil {
		return nil, err
	}
	return NewUserData(u), nil
}

func (us *userService) Update(ud *UserData) error {
	return us.UserRepository.Update(newUserFromData(ud))
}

func (us *userService) Delete(id uint32) error {
	return us.UserRepository.Delete(id)
}

func newUserFromData(data *UserData) *user.User {
	return user.NewUser(data.ID, data.Name, data.CreatedAt, data.UpdatedAt)
}
