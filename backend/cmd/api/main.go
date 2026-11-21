// Command api starts the TaskFlow REST API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow-backend/internal/config"
	"taskflow-backend/internal/handlers"
	"taskflow-backend/internal/middleware"
	"taskflow-backend/internal/repositories"
	"taskflow-backend/internal/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

// run wires all the layers. Returning an error (instead of calling os.Exit deep
// inside) lets the deferred cleanups run.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Readable logs in development, JSON logs in production.
	var logHandler slog.Handler = slog.NewTextHandler(os.Stdout, nil)
	if cfg.IsProduction() {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	}
	logger := slog.New(logHandler)

	// Cancelled on Ctrl+C or SIGTERM (Render sends SIGTERM when redeploying).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Layers: database -> repository -> service -> handler.
	client, db, err := config.ConnectDB(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return err
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(closeCtx); err != nil {
			logger.Error("closing mongodb", "error", err)
		}
	}()
	logger.Info("connected to mongodb", "database", cfg.MongoDatabase)

	taskRepository := repositories.NewTaskRepository(db)
	if err := taskRepository.EnsureIndexes(ctx); err != nil {
		return err
	}
	taskService := services.NewTaskService(taskRepository)

	// Routes.
	mux := http.NewServeMux()
	handlers.NewTaskHandler(taskService, logger).Register(mux)
	handlers.NewHealthHandler(taskRepository).Register(mux)

	// Middlewares: the first one is the outermost.
	handler := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logging(logger),
		middleware.Recover(logger),
		middleware.CORS(cfg.AllowedOrigins),
	)

	// IMPORTANT: the server must receive `handler` (the mux wrapped by the
	// middlewares). Passing nil would use Go's default mux and ignore all routes.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second, // protects against slowloris attacks
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr, "env", cfg.Env, "allowed_origins", cfg.AllowedOrigins)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		logger.Info("shutting down gracefully")
	}

	// Stop accepting connections and let in-flight requests finish (10 s max).
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
