package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var Db *sql.DB

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

	Db, err = sql.Open("sqlite", dbFileName)
	if err != nil {
		return err
	}
	defer Db.Close()

	if newDb {
		_, err := Db.Exec(`CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title VARCHAR(256) NOT NULL DEFAULT "new task",
			comment TEXT,
			repeat VARCHAR(128) NOT NULL DEFAULT "once"
		)`)
		if err != nil {
			return err
		}
		_, err = Db.Exec(`CREATE INDEX date_scheduler ON scheduler (date)`)
		if err != nil {
			return err
		}
	}

	return nil
}
