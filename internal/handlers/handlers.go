package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
	"github.com/mrScorpio/finalTask/internal/nextdate"
)

type jsonError struct {
	ErrText string `json:"error"`
}

type taskResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func NextDateHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		curTmStr := req.FormValue("now")

		var curTm time.Time
		var err error

		if curTmStr == "" {
			curTm = time.Now()
		} else {
			curTm, err = time.Parse(nextdate.TmFormat, curTmStr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		res, err := nextdate.NextDate(curTm, req.FormValue("date"), req.FormValue("repeat"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Write([]byte(res))
	}
}

func TaskHandler(w http.ResponseWriter, req *http.Request) {
	var task db.Task
	var buf bytes.Buffer

	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}

		if task.Title == "" {
			writeJson(w, jsonError{ErrText: "no title"})
			return
		}

		if err := nextdate.CheckDate(&task); err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
	}

	switch req.Method {
	case http.MethodPost:

		id, err := db.AddTask(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		task.Id = int(id)
		writeJson(w, task)

	case http.MethodGet:
		id, err := strconv.Atoi(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		task, err := db.GetTask(id)
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, task)

	case http.MethodPut:
		err := db.UpdTask(&task)
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)

	case http.MethodDelete:
		err := db.DelTask(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)
	}

}

func TasksHandler(w http.ResponseWriter, req *http.Request) {
	tasks, err := db.Tasks(11)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	writeJson(w, taskResp{Tasks: tasks})
}

func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func TaskDoneHandler(w http.ResponseWriter, req *http.Request) {

	id, err := strconv.Atoi(req.FormValue("id"))
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}

	if task.Repeat == "" {
		err := db.DelTask(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)
		return
	}
	tm, err := time.Parse(nextdate.TmFormat, task.Date)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	nxtdt, err := nextdate.NextDate(tm, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	if err := db.UpDateTask(nxtdt, req.FormValue("id")); err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	writeJson(w, w)
}
