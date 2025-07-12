// пакет для работы с БД
package db

// структура записи в планировщике
type Task struct {
	Id      int    `json:"id,string"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
