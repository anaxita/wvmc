package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/anaxita/wvmc/internal/app"
	"github.com/anaxita/wvmc/internal/wvmc/cache"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/dal"
	"github.com/anaxita/wvmc/internal/wvmc/notice"
	"github.com/anaxita/wvmc/internal/wvmc/server"
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

	repository := dal.New(db)
	cacheService := cache.NewCacheService()

	serviceServer := control.NewServerService(cacheService)
	noticeService := notice.NewNoticeService()
	s := server.New(repository, serviceServer, noticeService)

	go func() {
		s.UpdateAllServersInfo()(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)

			_, err := serviceServer.GetServersDataForAdmins()
			if err != nil {
				log.Println("update cache servers: ", err)
			}
		}
	}()

	if err = s.Start(); err != nil {
		log.Fatal("Ошибка запуска сервер", err)
	}
}
