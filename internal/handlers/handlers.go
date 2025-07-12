// пакет с хэндлерами хттп-запросов
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mrScorpio/finalTask/internal/db"
	"github.com/mrScorpio/finalTask/internal/nextdate"
)

// структура для вывода текста ошибки в джисоне
type jsonError struct {
	ErrText string `json:"error"`
}

// структура для приема пароля в джисоне
type jsonPass struct {
	Password string `json:"password"`
}

// структура для вывода токена в джисоне
type jsonToken struct {
	Token string `json:"token"`
}

// структура с указателями записей с оберткой в джисон
type taskResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// хэндлер проверки работы nextdate.NextDate(...)
func NextDateHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	curTmStr := req.FormValue("now")

	var curTm time.Time
	var err error

	if curTmStr == "" {
		curTm = time.Now()
	} else {
		curTm, err = time.Parse(db.TmFormat, curTmStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	res, err := nextdate.NextDate(curTm, req.FormValue("date"), req.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(res))
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
	searchStr := req.FormValue("search")
	limit := 66
	var tasks []*db.Task
	var err error
	if len(searchStr) > 0 {
		tasks, err = db.TasksSearchStr(limit, searchStr)
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
	} else {
		tasks, err = db.Tasks(limit)
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
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

	if _, ok := data.(jsonError); ok {
		http.Error(w, "", http.StatusBadRequest)
	}

	w.Write(resp)
}

// хэндлер обработки запроса о выполнении задачи
func TaskDoneHandler(w http.ResponseWriter, req *http.Request) {
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
	tm, err := time.Parse(db.TmFormat, task.Date)
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

// функция проверки пароля
func ChkPass(w http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	myPass := os.Getenv("TODO_PASSWORD")
	// проверили что пароль задан
	if len(myPass) < 1 {
		return
	}
	// если задан, зачитали содержимое формы
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pass := jsonPass{""}
	// пробуем десериализовать в пароль
	if err := json.Unmarshal(buf.Bytes(), &pass); err != nil {
		writeJson(w, jsonError{ErrText: err.Error()})
		return
	}

	if len(pass.Password) < 1 {
		writeJson(w, jsonError{ErrText: "unauthorised access prohibited"})
		return
	}
	if pass.Password == myPass {
		jwtToken := jwt.New(jwt.SigningMethodHS256)

		// получаем подписанный токен
		signedToken, err := jwtToken.SignedString([]byte(pass.Password))
		if err != nil {
			writeJson(w, jsonError{ErrText: err.Error()})
			return
		}
		// отвечаем токеном
		writeJson(w, jsonToken{Token: signedToken})
		return
	}
	writeJson(w, jsonError{ErrText: "wrong password"})
}

// функция аутентификации
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwtSigned string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtSigned = cookie.Value
			} else {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			// здесь код для валидации и проверки JWT-токена
			jwtToken, jwtErr := jwt.Parse(jwtSigned, func(t *jwt.Token) (interface{}, error) {
				// секретный ключ для всех токенов одинаковый, поэтому просто возвращаем его
				return []byte(pass), nil
			})

			if !jwtToken.Valid || jwtErr != nil {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
