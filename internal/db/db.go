package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type Task struct {
	Id      int    `json:"id,string"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

var db *sql.DB

func Init(dbFileName string) error {
	_, err := os.Stat(dbFileName)
	var newDb bool
	if err != nil {
		dbFile, err := os.OpenFile(dbFileName, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		dbFile.Close()
		newDb = true
	}

	db, err = sql.Open("sqlite", dbFileName)
	if err != nil {
		return err
	}

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

func CloseDb() {
	if db != nil {
		db.Close()
	}
}

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

func Tasks(limit int) ([]*Task, error) {
	tasks := make([]*Task, 0, limit)

	rows, err := db.Query("SELECT id,date,title,comment,repeat FROM scheduler")
	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func GetTask(id int) (*Task, error) {
	var task Task
	task.Id = id
	row := db.QueryRow("SELECT date,title,comment,repeat FROM scheduler WHERE id=:id", sql.Named("id", id))
	return &task, row.Scan(&task.Date, &task.Title, &task.Comment, &task.Repeat)
}

func UpdTask(task *Task) error {

	res, err := db.Exec("UPDATE scheduler SET date=:date,title=:title,comment=:comment,repeat=:repeat WHERE id=:id",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.Id))
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}

	return nil
}

func DelTask(id string) error {
	res, err := db.Exec("DELETE FROM scheduler WHERE id=:id", sql.Named("id", id))
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}

	return nil
}

func UpDateTask(next string, id string) error {
	res, err := db.Exec("UPDATE scheduler SET date=:date WHERE id=:id",
		sql.Named("date", next),
		sql.Named("id", id))
	if err != nil {
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return fmt.Errorf("incorrect id")
	}
	return nil
}
