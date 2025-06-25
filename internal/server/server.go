package server

import (
	"log"
	"net/http"
	"time"
)

type MyServ struct {
	Serv  http.Server
	Loger log.Logger
}

func NewServer(loger log.Logger, port string) *MyServ {
	mux := http.NewServeMux()

	//	mux.HandleFunc("/", handlers.HandleMain)
	mux.Handle("/", http.FileServer(http.Dir("./web")))

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
