package main

import (
	"flag"
	"fmt"
	"github.com/anaxita/wvmc/internal/wvmc/cache"
	"github.com/anaxita/wvmc/internal/wvmc/notice"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/server"
	"github.com/anaxita/wvmc/internal/wvmc/store"
)

var envPath string

func main() {
	flag.StringVar(&envPath, "e", ".env", "path to .env")
	flag.Parse()

	err := godotenv.Load(envPath)
	if err != nil {
		f, _ := os.OpenFile("./errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0200)
		defer f.Close()
		f.WriteString(fmt.Sprintln(time.Now().Format("02.01.2006 15:04:05"), err))
		log.Fatal("[FATAL] Cannot find env file")
	}

	err = logit.New(os.Getenv("LOG"))
	if err != nil {
		log.Fatal("Не удалось запустить логгер", err)
	}
	defer logit.Close()

	db, err := store.Connect(os.Getenv("DB_TYPE"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err != nil {
		logit.Fatal("Ошибка соединения с БД:", err)
	}
	defer db.Close()

	err = store.Migrate(db)
	if err != nil {
		logit.Fatal("Ошибка миграции", err)
	}

	repository := store.New(db)
	cacheService := cache.NewCacheService()

	serviceServer := control.NewServerService(new(control.Command), cacheService)
	noticeService := notice.NewNoticeService()
	s := server.New(repository, serviceServer, noticeService)

	go func() {
		s.UpdateAllServersInfo()(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)

			_, err := serviceServer.GetServersDataForAdmins()
			if err != nil {
				logit.Log("update cache servers: ", err)
			}
		}
	}()

	if err = s.Start(); err != nil {
		logit.Fatal("Ошибка запуска сервер", err)
	}
}
