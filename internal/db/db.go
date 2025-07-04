// пакет для работы с БД
package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

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
			return err
		}
		dbFile.Close()
		newDb = true
	}
	// подключаемся к БД
	db, err = sql.Open("sqlite", dbFileName)
	if err != nil {
		return err
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
			return err
		}
		_, err = db.Exec(`CREATE INDEX date_scheduler ON scheduler (date)`)
		if err != nil {
			return err
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
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// функция чтения заданного количества записей из базы
func Tasks(limit int) ([]*Task, error) {
	// слайс, в который читаем
	tasks := make([]*Task, 0, limit)
	// эскуэль запрос
	rows, err := db.Query("SELECT id,date,title,comment,repeat FROM scheduler ORDER BY date LIMIT :limit",
		sql.Named("limit", limit))
	if err != nil {
		return tasks, err
	}
	defer rows.Close()
	// бежим по строкам
	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, err
		}
		//и заполняем слайс
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// функция запроса записи БД по айди
func GetTask(id string) (*Task, error) {
	var task Task
	var err error
	task.Id, err = strconv.Atoi(id)
	if err != nil {
		return &task, err
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
		return err
	}
	// проверили количество измененных
	num, err := res.RowsAffected()
	if err != nil {
		return err
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
		return err
	}
	//проверяем, что удалилась
	num, err := res.RowsAffected()
	if err != nil {
		return err
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
		return err
	}
	// проверяем, запись нашлась
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}
	return nil
}
