package server

import (
	"log"
	"net/http"
	"time"

	"github.com/mrScorpio/finalTask/internal/handlers"
)

type MyServ struct {
	Serv  http.Server
	Loger log.Logger
}

func NewServer(loger log.Logger, port string) *MyServ {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	mux.HandleFunc("/api/task", handlers.TaskHandler)
	mux.HandleFunc("/api/tasks", handlers.TasksHandler)
	mux.HandleFunc("/api/task/done", handlers.TaskDoneHandler)

	rdTmOut := 5 * time.Second
	wrTmOut := 10 * time.Second
	idleTmOut := 15 * time.Second

	serv := http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ErrorLog:     &loger,
		ReadTimeout:  rdTmOut,
		WriteTimeout: wrTmOut,
		IdleTimeout:  idleTmOut,
	}

	myserv := MyServ{Loger: loger, Serv: serv}

	return &myserv
}
