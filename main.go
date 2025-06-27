package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mrScorpio/finalTask/internal/db"
	"github.com/mrScorpio/finalTask/internal/server"
	"github.com/mrScorpio/finalTask/tests"
)

func main() {
	logFile, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	myLog := log.New(logFile, `http-server`, log.LstdFlags|log.Lshortfile)

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = fmt.Sprint(tests.Port)
	}

	myServ := server.NewServer(*myLog, port)

	err = db.Init("scheduler.db")
	if err != nil {
		myLog.Fatal(err.Error())
	}

	err = myServ.Serv.ListenAndServe()
	if err != nil {
		myLog.Fatal(err.Error())
	}

}
