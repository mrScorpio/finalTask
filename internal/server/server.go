// пакет для настройки хттп-сервера
package server

import (
	"log"
	"net/http"
	"time"

	"github.com/mrScorpio/finalTask/internal/handlers"
)

// структура сервера с прикрученным логом
type MyServ struct {
	Serv  http.Server
	Loger log.Logger
}

// функция создания нового экзепляра сервера с логом
func NewServer(loger log.Logger, port string) *MyServ {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	mux.HandleFunc("/api/task", handlers.Auth(handlers.TaskHandler))
	mux.HandleFunc("/api/tasks", handlers.Auth(handlers.TasksHandler))
	mux.HandleFunc("/api/task/done", handlers.Auth(handlers.TaskDoneHandler))
	mux.HandleFunc("/api/signin", handlers.ChkPass)

	serv := http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ErrorLog:     &loger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	myserv := MyServ{Loger: loger, Serv: serv}

	return &myserv
}
