// пакет для работы с БД
package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

// формат хранения даты в строке
const TmFormat string = "20060102"

var db *sql.DB

// функция инициализации БД
func Init(dbFileName string) error {
	// проверяем наличие файла с БД
	_, err := os.Stat(dbFileName)
	var newDb bool
	// если его нет, то создаем
	if err != nil {
		dbFile, err := os.OpenFile(dbFileName, os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("can't open db-file: %w", err)
		}
		dbFile.Close()
		newDb = true
	}
	// подключаемся к БД
	db, err = sql.Open("sqlite", dbFileName)
	if err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}
	//если создали файл, то создаем в нем таблицу и индекс
	if newDb {
		_, err := db.Exec(`CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title VARCHAR(256) NOT NULL DEFAULT "задача",
			comment TEXT NOT NULL DEFAULT "",
			repeat VARCHAR(128) NOT NULL DEFAULT ""
		)`)
		if err != nil {
			return fmt.Errorf("error while creating table: %w", err)
		}
		_, err = db.Exec(`CREATE INDEX date_scheduler ON scheduler (date)`)
		if err != nil {
			return fmt.Errorf("error while creating index on scheduler: %w", err)
		}
	}

	return nil
}

// функция для зарытия БД из других пакетов
func CloseDb() {
	if db != nil {
		db.Close()
	}
}

// функция добавления новой записи в БД
func AddTask(task *Task) (int64, error) {
	var id int64
	res, err := db.Exec("INSERT INTO scheduler (date,title,comment,repeat) VALUES (:date,:title,:comment,:repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, fmt.Errorf("can't insert new task: %w", err)
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("can't get index of inserted task: %w", err)
	}
	return id, nil
}

// функция чтения заданного количества записей из базы
func Tasks(limit int) ([]*Task, error) {
	// слайс, в который читаем
	tasks := make([]*Task, 0, limit)
	// эскуэль запрос
	rows, err := db.Query("SELECT id,date,title,comment,repeat FROM scheduler ORDER BY date LIMIT :limit",
		sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error while SELECT query: %w", err)
	}
	defer rows.Close()
	// бежим по строкам
	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("error while scan table: %w", err)
		}
		//и заполняем слайс
		tasks = append(tasks, &task)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("some error in cursor: %w", err)
	}

	return tasks, nil
}

// функция запроса записи БД по айди
func GetTask(id string) (*Task, error) {
	var task Task
	var err error
	task.Id, err = strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("can't convert ID to int: %w", err)
	}
	row := db.QueryRow("SELECT date,title,comment,repeat FROM scheduler WHERE id=:id", sql.Named("id", id))
	return &task, row.Scan(&task.Date, &task.Title, &task.Comment, &task.Repeat)
}

// функция изменения всех полей записи БД по айди
func UpdTask(task *Task) error {
	// запросили
	res, err := db.Exec("UPDATE scheduler SET date=:date,title=:title,comment=:comment,repeat=:repeat WHERE id=:id",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return fmt.Errorf("can't update task: %w", err)
	}
	// проверили количество измененных
	num, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't check updated rows: %w", err)
	}
	// если их нет, то ошибка
	if num == 0 {
		return fmt.Errorf("incorrect id")
	}

	return nil
}

// функция удаления записи по айди
func DelTask(id string) error {
	res, err := db.Exec("DELETE FROM scheduler WHERE id=:id", sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("can't delete task: %w", err)
	}
	//проверяем, что удалилась
	num, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't check deleted rows: %w", err)
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}

	return nil
}

// функция изменения поля с датой записи
func UpDateTask(next string, id string) error {
	res, err := db.Exec("UPDATE scheduler SET date=:date WHERE id=:id",
		sql.Named("date", next),
		sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("can't update task date: %w", err)
	}
	// проверяем, запись нашлась
	num, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't check updated task date: %w", err)
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}
	return nil
}

// функция поиска записей в базе по словам в заголовке и коментах или дате формата 02.01.2006
func TasksSearchStr(limit int, str string) ([]*Task, error) {
	// слайс, в который читаем
	tasks := make([]*Task, 0, limit)
	query := "SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit"
	search := "%" + str + "%"
	// если задана дата в нужном формате, то меняем запрос
	date, err := time.Parse("02.01.2006", str)
	if err == nil {
		search = date.Format(TmFormat)
		query = "SELECT * FROM scheduler WHERE date = :search ORDER BY date LIMIT :limit"
	}
	// эскуэль запрос
	rows, err := db.Query(query, sql.Named("search", search), sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error while query for search: %w", err)
	}
	defer rows.Close()
	// бежим по строкам
	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("error while scan for search: %w", err)
		}
		//и заполняем слайс
		tasks = append(tasks, &task)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("some error in cursor: %w", err)
	}

	return tasks, nil
}
