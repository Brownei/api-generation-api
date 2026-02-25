package services

import (
	"errors"
	"fmt"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/types"
	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	cfg *config.AppConfig
}

func NewUserService(db *gorm.DB, cfg *config.AppConfig) *UserService {
	return &UserService{db: db, cfg: cfg}
}

func (u *UserService) CreateAUser(email, password string) (*db.User, error) {
	var newUser db.User

	fmt.Printf("Creating a user")
	user := db.User{
		Email:    email,
		Password: password,
	}

	if err := u.db.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}

		return nil, err
	}

	return &newUser, nil
}

func (u *UserService) FindThisUser(email string) (*db.User, error) {
	var newUser db.User
	if err := u.db.Where("email = ?", email).First(&newUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}

		return nil, err
	}

	return &newUser, nil
}
