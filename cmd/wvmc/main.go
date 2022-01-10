package main

import (
	"fmt"
	"github.com/anaxita/wvmc/internal/wvmc/cache"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/anaxita/logit"
	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/server"
	"github.com/anaxita/wvmc/internal/wvmc/store"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env_prod")
	if err != nil {
		f, _ := os.OpenFile("./errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0200)
		defer f.Close()
		f.WriteString(fmt.Sprintln(time.Now().Format("02.01.2006 15:04:05"), err))
		log.Fatal("[FATAL] Cannot find env file")
	}
}

func main() {
	err := logit.New(os.Getenv("LOG"))
	if err != nil {
		log.Fatal("Не удалось запустить логгер", err)
	}
	defer logit.Close()

	db, err := store.Connect(os.Getenv("DB_TYPE"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_ADDR"), os.Getenv("DB_NAME"))
	if err != nil {
		logit.Fatal("Ошибка соединения с БД:", err)
	}
	defer db.Close()

	err = store.Migrate(db)
	if err != nil {
		logit.Fatal("Ошибка миграции", err)
	}

	store := store.New(db)
	cacheService := cache.NewCacheService()

	serviceServer := control.NewServerService(new(control.Command), cacheService)
	s := server.New(store, serviceServer)

	go func() {
		s.UpdateAllServersInfo()(httptest.NewRecorder(), &http.Request{})

		for {
			time.Sleep(time.Minute * 1)
			serviceServer.GetServersDataForAdmins()
		}
	}()

	if err = s.Start(); err != nil {
		logit.Fatal("Ошибка запуска сервер", err)
	}
}
