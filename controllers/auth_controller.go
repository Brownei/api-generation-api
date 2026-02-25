package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Brownei/api-generation-api/dto"
	"github.com/Brownei/api-generation-api/services"
	"github.com/Brownei/api-generation-api/types"
	"github.com/Brownei/api-generation-api/utils"
	"go.uber.org/zap"
)

type AuthController struct {
	userService *services.UserService
	authService *services.AuthService
	logger      *zap.SugaredLogger
}

func NewAuthController(userService *services.UserService, authService *services.AuthService, logger *zap.SugaredLogger) *AuthController {
	return &AuthController{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (a *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var authDto dto.AuthDto
	if err := utils.ParseJSON(r, &authDto); err != nil {
		utils.WriteError(w, 500, errors.New("Cannot parse the data correctly"))
		return
	}

	existingUser, err := a.userService.FindThisUser(authDto.Email)
	if err != nil {
		if errors.Is(err, types.ErrUserNotFound) {
			utils.WriteError(w, 404, err)
			return
		}

		utils.WriteError(w, 409, err)
		return
	}

	isPasswordCorrect := a.authService.CheckPassword(authDto.Pasword, existingUser.Password)
	if isPasswordCorrect == false {
		utils.WriteError(w, 409, types.ErrInvalidCredentials)
		return
	}

	token, err := a.authService.GenerateToken(existingUser.ID, existingUser.Email)
	if err != nil {
		utils.WriteError(w, 409, err)
		return
	}

	// cookie := &http.Cookie{
	// 	Name:     "auth_user_token",
	// 	Value:    token,
	// 	Expires:  time.Now().Add(24 * time.Hour), // 24 hours
	// 	HttpOnly: true,                           // Prevent JavaScript access
	// 	Secure:   true,                           // Only send over HTTPS
	// 	Path:     "/",
	// 	SameSite: http.SameSiteStrictMode,
	// }
	//
	// http.SetCookie(w, cookie)
	//
	utils.WriteJSON(w, 200, []byte(token))
}

func (a *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var authDto dto.AuthDto
	if err := utils.ParseJSON(r, &authDto); err != nil {
		utils.WriteError(w, 500, errors.New("Cannot parse the data correctly"))
		return
	}

	existingUser, err := a.userService.FindThisUser(authDto.Email)
	fmt.Printf("New direct user %v", existingUser)

	if err != nil {
		if errors.Is(err, types.ErrUserNotFound) {
			hashedPassword, err := a.authService.HashPassword(authDto.Pasword)
			if err != nil {
				utils.WriteError(w, 409, err)
				return
			}

			newUser, err := a.userService.CreateAUser(authDto.Email, hashedPassword)
			if err != nil {
				if errors.Is(err, types.ErrUserNotFound) {
					utils.WriteError(w, 404, err)
					return
				}

				utils.WriteError(w, 409, err)
				return
			}

			fmt.Printf("New direct user %v", &newUser)
			isPasswordCorrect := a.authService.CheckPassword(authDto.Pasword, newUser.Password)
			fmt.Printf("Is it correct? %v", &isPasswordCorrect)
			if isPasswordCorrect == false {
				utils.WriteError(w, 409, types.ErrInvalidCredentials)
				return
			}

			token, err := a.authService.GenerateToken(newUser.ID, newUser.Email)
			if err != nil {
				utils.WriteError(w, 409, err)
				return
			}

			// cookie := &http.Cookie{
			// 	Name:     "auth_user_token",
			// 	Value:    token,
			// 	Expires:  time.Now().Add(24 * time.Hour), // 24 hours
			// 	HttpOnly: true,                           // Prevent JavaScript access
			// 	Secure:   true,                           // Only send over HTTPS
			// 	Path:     "/",
			// 	SameSite: http.SameSiteStrictMode,
			// }
			//
			// http.SetCookie(w, cookie)

			utils.WriteJSON(w, 200, []byte(token))
			return
		}

		utils.WriteError(w, 500, errors.New("Something happened"))
		return
	}

	utils.WriteJSON(w, 304, []byte("This user has already been created!"))
}
