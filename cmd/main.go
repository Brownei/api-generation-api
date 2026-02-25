package main

import (
	"log"

	"github.com/Brownei/api-generation-api/cmd/api"
	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/db"
	"github.com/Brownei/api-generation-api/store"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	sugarLogger := logger.Sugar()
	cfg := config.LoadAppConfig()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	store := store.NewStore(database, cfg, sugarLogger)
	application := api.NewApplication(sugarLogger, cfg, database, store)

	if err := application.Run(); err != nil {
		sugarLogger.Fatalf("Failed to run the application: %v", err)
	}

}
