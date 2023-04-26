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

	err = app.UpMigrations(db.DB, c.DB.Name, c.DB.MigrationsPath)
	if err != nil {
		l.Fatalf("failed to run migrations: %v", err)
	}

	userRepo := dal.NewUserRepository(db)
	serverRepo := dal.NewServerRepository(db)

	userService := service.NewUserService(userRepo)
	serverService := service.NewServerService(serverRepo)
	authService := service.NewAuthService(userRepo)
	cacheService := dal.NewCache()
	controlService := service.NewControlService(cacheService)
	notifier := notice.NewNoticeService()

	userHandler := api.NewUserHandler(userService, serverService)
	serverHandler := api.NewServerHandler(serverService, controlService, notifier)
	authHandler := api.NewAuthHandler(userService, authService)
	mw := api.NewMiddleware(userService, serverService)

	s := api.NewServer(c.HTTPPort, userHandler, authHandler, serverHandler, mw)

	go func() {
		serverHandler.UpdateAllServersInfo()(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)

			_, err := controlService.GetServersDataForAdmins()
			if err != nil {
				log.Println("update cache servers: ", err)
			}
		}
	}()

	if err = s.ListenAndServe(); err != nil {
		log.Fatal("Ошибка запуска сервер", err)
	}
}
