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

	"github.com/anaxita/wvmc/internal/api"
	"github.com/anaxita/wvmc/internal/app"
	"github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/scheduler"
	"github.com/anaxita/wvmc/internal/service"
	"go.uber.org/zap"
)

func main() {
	c, err := app.NewConfig()
	if err != nil {
		log.Fatalf(`{"msg": "failed to load config: %v"}`, err)
	}

	l, err := app.NewLogger(c.LogFile)
	if err != nil {
		log.Fatalf(`{"msg": "failed to create logger: %v"}`, err)
	}
	defer l.Sync()

	db, err := app.NewSQLite3Client(c.DB)
	if err != nil {
		l.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	err = app.UpMigrations(db.DB, c.DB.Name, "migrations")
	if err != nil {
		l.Fatalf("failed to run migrations: %s", err)
	}

	userRepo := dal.NewUserRepository(db)
	serverRepo := dal.NewServerRepository(db)

	userService := service.NewUserService(userRepo)
	controlService := service.NewControlService()
	serverService := service.NewServerService(serverRepo, controlService)
	authService := service.NewAuthService(userRepo)
	notifierService := notice.NewNoticeService()

	userHandler := api.NewUserHandler(l, userService, serverService)
	serverHandler := api.NewServerHandler(l, serverService, controlService, notifierService)
	authHandler := api.NewAuthHandler(l, userService, authService)
	mw := api.NewMiddleware(l, userService, serverService)

	server := api.NewServer(c.HTTPPort, userHandler, authHandler, serverHandler, mw)
	go func() {
		l.Info("starting http server on port: ", c.HTTPPort)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Panicw("failed to start server", zap.Error(err))
		}
	}()

	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	defer schedulerCancel()

	// Scheduler.
	{
		scheduler.NewScheduler(l).
			AddJob("update servers info", time.Minute, true, serverService.LoadServersFromHVs).
			Start(schedulerCtx)
	}

	// Graceful shutdown.
	{
		notifyCh := make(chan os.Signal, 1)
		signal.Notify(notifyCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		s := <-notifyCh
		l.Info("shutting down server due to signal: ", s)
		schedulerCancel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		l.Info("shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			l.Panicw("failed to shutdown server", zap.Error(err))
		}

		l.Info("server stopped")
	}
}
