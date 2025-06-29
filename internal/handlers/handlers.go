package handlers

import (
	"net/http"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
	"github.com/mrScorpio/finalTask/internal/nextdate"
)

func NextDateHandler(w http.ResponseWriter, req *http.Request) {
	var curTm time.Time
	var err error
	if req.Method == http.MethodPost {
		curTmStr := req.FormValue("now")
		if curTmStr == "" {
			curTm = time.Now()
		} else {
			curTm, err = time.Parse("20060102", curTmStr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		db.TaskItem.Date, err = nextdate.NextDate(curTm, req.FormValue("date"), req.FormValue("repeat"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(db.TaskItem.Date))
	}
}
