package controllers

import (
	"errors"
	"net/http"

	"github.com/Brownei/api-generation-api/dto"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/types"
	"github.com/Brownei/api-generation-api/utils"
	"go.uber.org/zap"
)

type UserController struct {
	userService *services.UserService
	authService *services.AuthService
	logger      *zap.SugaredLogger
}

func NewUserController(userService *services.UserService, authService *services.AuthService, logger *zap.SugaredLogger) *UserController {
	return &UserController{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (u *UserController) CreateANewUser(w http.ResponseWriter, r *http.Request) {
	var userDto dto.UserDto
	if err := utils.ParseJSON(r, &userDto); err != nil {
		utils.WriteError(w, 409, errors.New("Cannot parse the data correctly"))
	}

	hashedPassword, err := u.authService.HashPassword(userDto.Pasword)
	if err != nil {
		utils.WriteError(w, 409, errors.New("Error hashing the password"))
	}

	newUser, err := u.userService.CreateAUser(userDto.Email, hashedPassword)
	if err != nil {
		utils.WriteError(w, 409, errors.New("Error hashing the password"))
	}

	utils.WriteJSON(w, 201, &newUser)
}

func (u *UserController) FindAUser(w http.ResponseWriter, r *http.Request) {
	var userDto dto.UserEmail
	if err := utils.ParseJSON(r, &userDto); err != nil {
		utils.WriteError(w, 409, errors.New("Cannot parse the data correctly"))
	}

	existingUser, err := u.userService.FindThisUser(userDto.Email)
	if err != nil {
		if errors.Is(err, types.ErrUserNotFound) {
			utils.WriteError(w, 404, err)
		}

		utils.WriteError(w, 500, err)
	}

	utils.WriteJSON(w, 200, &existingUser)
}
