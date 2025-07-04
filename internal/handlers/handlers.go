// пакет с хэндлерами хттп-запросов
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
	"github.com/mrScorpio/finalTask/internal/nextdate"
)

// структура для вывода текста ошибки в джисоне
type jsonError struct {
	ErrText string `json:"error"`
}

// структура с указателями записей с оберткой в джисон
type taskResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// хэндлер проверки работы nextdate.NextDate(...)
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

// хэндлер обработки задачи
func TaskHandler(w http.ResponseWriter, req *http.Request) {
	var task db.Task
	var buf bytes.Buffer
	// общие действия для пост- и пут-запросов
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		// зачитали содержимое
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// пробуем десериализовать в задачу
		if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		// проверили заголовок
		if task.Title == "" {
			writeJson(w, jsonError{ErrText: "no title"})
			return
		}
		// проверили остальные поля
		if err := nextdate.CheckDate(&task); err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
	}

	switch req.Method {
	case http.MethodPost:
		// если пост-, то добавляем задачу в базу
		id, err := db.AddTask(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		task.Id = int(id)
		writeJson(w, task)

	case http.MethodGet:
		/*
			id, err := strconv.Atoi(req.FormValue("id"))
			if err != nil {
				writeJson(w, jsonError{ErrText: err.Error()})
				return
			}
		*/
		//если гет-, то достаем из базы по айди
		task, err := db.GetTask(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, task)

	case http.MethodPut:
		// если пут-, то изменяем запись в базе
		err := db.UpdTask(&task)
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)

	case http.MethodDelete:
		// если делит-, то удаляем из базы
		err := db.DelTask(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)
	}

}

// хэндлер вывода списка задач из базы в джисон
func TasksHandler(w http.ResponseWriter, req *http.Request) {
	tasks, err := db.Tasks(66)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	writeJson(w, taskResp{Tasks: tasks})
}

// функция вывода результатов работы хэндлеров в джисон-формате
func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

// хэндлер обработки запроса о выполнении задачи
func TaskDoneHandler(w http.ResponseWriter, req *http.Request) {
	/*
		id, err := strconv.Atoi(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
	*/
	// зачитали задачу из базы
	task, err := db.GetTask(req.FormValue("id"))
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	// если нет правила повторения, то удаляем
	if task.Repeat == "" {
		err := db.DelTask(req.FormValue("id"))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		writeJson(w, w)
		return
	}
	// если правило есть, то анализируем дату
	tm, err := time.Parse(nextdate.TmFormat, task.Date)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	// рассчитываем новую дату
	nxtdt, err := nextdate.NextDate(tm, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	// и обновляем ее в базе
	if err := db.UpDateTask(nxtdt, req.FormValue("id")); err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}
	writeJson(w, w)
}
