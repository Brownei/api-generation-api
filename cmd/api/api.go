package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Brownei/api-generation-api/config"
	"github.com/Brownei/api-generation-api/store"
	"github.com/Brownei/api-generation-api/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Application struct {
	addr   string
	logger *zap.SugaredLogger
	store  *store.Store
	cfg    *config.AppConfig
	db     *gorm.DB
}

func NewApplication(logger *zap.SugaredLogger, cfg *config.AppConfig, db *gorm.DB, store *store.Store) *Application {
	return &Application{
		addr:   ":8080",
		logger: logger,
		cfg:    cfg,
		db:     db,
		store:  store,
	}
}

func (a *Application) Run() error {
	r := chi.NewRouter()

	server := &http.Server{
		Addr:         a.addr,
		Handler:      enableCORS(r),
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 5,
	}

	authController := a.store.AuthController
	apiKeyController := a.store.APIKeyController
	userController := a.store.UserController

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))).ServeHTTP(w, r)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, []byte("Welcome to Brownson Esiti's Submission"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Post("/api/auth/register", authController.Register)
		r.Post("/api/auth/login", authController.Login)

		r.Route("/api", func(r chi.Router) {
			r.Use(AuthMiddleware(a.cfg.JWTSecret))

			r.Get("/users/{id}", userController.FindAUser)

			r.Route("/api-key", func(r chi.Router) {
				r.Post("/", apiKeyController.CreateAPIKey)
				r.Get("/", apiKeyController.ListAPIKeys)
				r.Get("/{id}", apiKeyController.RevokeAPIKey)
			})
		})
	})

	// Run the server in a goroutine so it doesn't block
	go func() {
		log.Printf("Running currently on %s", ":8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Set up channel on which to send signal notifications.
	// We’ll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM.
	// SIGKILL, SIGQUIT will not be caught.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal.
	<-stop

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown.
	log.Printf("Shutting down server gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("Server exiting")
	return nil
}
