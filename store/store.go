package store

import (
	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/controllers"
	"github.com/Brownei/api-generation-api/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

type Store struct {
	APIKeyController *controllers.APIKeyController
	UserController   *controllers.UserController
	AuthController   *controllers.AuthController
}

func NewStore(db *gorm.DB, cfg *config.AppConfig, logger *zap.SugaredLogger) *Store {
	apiKeyService := services.NewAPIKeyService(db)
	authService := services.NewAuthService(db, cfg)
	userService := services.NewUserService(db, cfg)

	return &Store{
		APIKeyController: controllers.NewAPIKeyController(apiKeyService, logger),
		UserController:   controllers.NewUserController(userService, authService, logger),
		AuthController:   controllers.NewAuthController(userService, authService, logger),
	}
}
