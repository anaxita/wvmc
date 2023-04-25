package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/anaxita/wvmc/internal/app"
	dal2 "github.com/anaxita/wvmc/internal/dal"
	"github.com/anaxita/wvmc/internal/notice"
	"github.com/anaxita/wvmc/internal/server"
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

	userRepo := dal2.NewUserRepository(db)
	serverRepo := dal2.NewServerRepository(db)

	userService := service.NewUserService(userRepo)
	serverService := service.NewServerService(serverRepo)
	authService := service.NewAuthService(userRepo)
	cacheService := dal2.NewCache()
	controlService := service.NewControlService(cacheService)
	noticeService := notice.NewNoticeService()

	s := server.New(controlService, noticeService, userService, serverService, authService)

	go func() {
		s.UpdateAllServersInfo()(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)

			_, err := controlService.GetServersDataForAdmins()
			if err != nil {
				log.Println("update cache servers: ", err)
			}
		}
	}()

	if err = s.Start(); err != nil {
		log.Fatal("Ошибка запуска сервер", err)
	}
}
