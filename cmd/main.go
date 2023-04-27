package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/anaxita/wvmc/internal/api"
	"github.com/anaxita/wvmc/internal/app"
	"github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/service"
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
		l.Fatalf("failed to run migrations: %v", err)
	}

	userRepo := dal.NewUserRepository(db)
	serverRepo := dal.NewServerRepository(db)
	cacheService := dal.NewCache()

	userService := service.NewUserService(userRepo)
	controlService := service.NewControlService(cacheService)
	serverService := service.NewServerService(serverRepo, controlService)
	authService := service.NewAuthService(userRepo)
	notifier := notice.NewNoticeService()

	userHandler := api.NewUserHandler(l, userService, serverService)
	serverHandler := api.NewServerHandler(l, serverService, controlService, notifier)
	authHandler := api.NewAuthHandler(l, userService, authService)
	mw := api.NewMiddleware(l, userService, serverService)

	s := api.NewServer(c.HTTPPort, userHandler, authHandler, serverHandler, mw)

	go func() {
		serverHandler.UpdateAllServersInfo(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)

			_, err := controlService.GetServersDataForAdmins()
			if err != nil {
				log.Println("update cache servers: ", err)
			}
		}
	}()

	if err = s.ListenAndServe(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
