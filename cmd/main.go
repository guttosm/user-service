package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guttosm/user-service/config"
	"github.com/guttosm/user-service/internal/app"
)

// startServer starts the HTTP server and listens for incoming requests.
//
// Parameters:
//   - router (http.Handler): The HTTP router to handle incoming requests.
//   - port (string): The port the server should listen on.
//
// Returns:
//   - *http.Server: The initialized server instance.
func startServer(router http.Handler, port string) *http.Server {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server running on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	return server
}

// gracefulShutdown gracefully shuts down the server when an interrupt signal is received.
//
// Parameters:
//   - ctx (context.Context): A shared context with timeout.
//   - server (*http.Server): The HTTP server instance to shut down.
//   - cleanup (func()): A callback to clean up resources (e.g., DB connections).
func gracefulShutdown(ctx context.Context, server *http.Server, cleanup func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cleanup()
	log.Println("Server exited gracefully")
}

// main is the entry point of the application.
// It loads the configuration, initializes the app,
// validates critical settings, and starts the HTTP server.
func main() {
	ctx := context.Background()

	config.LoadConfig()

	cfg := config.AppConfig
	if cfg.Server.Port == "" || cfg.Postgres.Host == "" || cfg.JWT.Secret == "" {
		log.Fatal("Missing required configuration. Check your .env or environment variables.")
	}

	router, cleanup, err := app.InitializeApp()
	if err != nil {
		log.Fatal("Error initializing application:", err)
	}

	server := startServer(router, cfg.Server.Port)
	gracefulShutdown(ctx, server, cleanup)
}
