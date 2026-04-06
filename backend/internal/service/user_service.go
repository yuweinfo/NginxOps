package service

import (
	"errors"
	"nginxops/internal/model"
	"nginxops/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) GetByUsername(username string) (*model.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) VerifyPassword(id uint, rawPassword string) bool {
	user, err := s.userRepo.FindByID(id)
	if err != nil || user == nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword))
	return err == nil
}

func (s *UserService) UpdateProfile(id uint, email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if email != "" {
		user.Email = email
	}

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
