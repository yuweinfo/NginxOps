package repository

import (
	"errors"
	"nginxops/internal/database"
	"nginxops/internal/model"
)

var ErrDatabaseNotConnected = errors.New("数据库未连接")

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	if database.DB == nil {
		return nil, ErrDatabaseNotConnected
	}
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	if database.DB == nil {
		return nil, ErrDatabaseNotConnected
	}
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *model.User) error {
	if database.DB == nil {
		return ErrDatabaseNotConnected
	}
	return database.DB.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	if database.DB == nil {
		return ErrDatabaseNotConnected
	}
	return database.DB.Save(user).Error
}

func (r *UserRepository) Count() (int64, error) {
	if database.DB == nil {
		return 0, ErrDatabaseNotConnected
	}
	var count int64
	if err := database.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
