// пакет расчета следующей даты при изменении записи
package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
)

// формат хранения даты в строке
const TmFormat string = "20060102"

// функция возвращает новую дату для задачи, принимает (текущее время, дата задачи, правило повторения)
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// если правило повторения пустое, ничего не делаем
	if repeat == "" {
		return "", nil
	}
	// преобразуем строку с датой в переменную типа тайм
	date, err := time.Parse(TmFormat, dstart)
	if err != nil {
		return "", err
	}
	// разбираем строку с правилом повторения в слайс
	rep := strings.Split(repeat, " ")
	// если только один символ и это не год, то плохое правило
	if rep[0] != "y" && len(rep) < 2 {
		return "", fmt.Errorf("wrong repeat format")
	}
	// добавляем год
	if rep[0] == "y" {
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				break
			}

		}
		return date.Format(TmFormat), nil
	}
	// добавляем дни
	if rep[0] == "d" {
		interval, err := strconv.Atoi(rep[1])
		if err != nil {
			return "", err
		}
		if interval > 400 {
			return "", fmt.Errorf("interval is too big")
		}
		for {
			date = date.AddDate(0, 0, interval)
			if date.After(now) {
				break
			}

		}
		return date.Format(TmFormat), nil
	}
	// дни недели
	if rep[0] == "w" {
		weekDays := strings.Split(rep[1], ",")

		for {
			date = date.AddDate(0, 0, 1)
			weekDayMatch := false
			for _, v := range weekDays {
				weekDayNum, err := strconv.Atoi(v)
				if err != nil {
					return "", err
				}
				if weekDayNum > 7 || weekDayNum < 0 {
					return "", fmt.Errorf("wrong weekday number")
				}
				if weekDayNum == 7 {
					weekDayNum = 0
				}
				if date.Weekday() == time.Weekday(weekDayNum) {
					weekDayMatch = true
					break
				}
			}
			if date.After(now) && weekDayMatch {
				break
			}

		}
		return date.Format(TmFormat), nil
	}

	// дни месяца
	if rep[0] == "m" {
		monthDays := strings.Split(rep[1], ",")

		monthMatch := true
		monthNums := make([]string, 0, 12)
		if len(rep) > 2 {
			monthNums = strings.Split(rep[2], ",")
			monthMatch = false
		}
		for {
			date = date.AddDate(0, 0, 1)
			monthDayMatch := false

			for _, v := range monthDays {
				monthDay, err := strconv.Atoi(v)

				if err != nil {
					return "", err
				}

				if monthDay > 31 || monthDay < -2 || monthDay == 0 {
					return "", fmt.Errorf("monthday number is bad")
				}

				if monthDay == -2 || monthDay == -1 {
					return "", fmt.Errorf("monthday number is bad")
				}

				if date.Day() == monthDay {
					monthDayMatch = true
					break
				}
			}

			for _, v := range monthNums {
				monthNum, err := strconv.Atoi(v)
				if err != nil {
					return "", err
				}
				if monthNum > 12 || monthNum < 1 {
					return "", fmt.Errorf("month num is wrong")
				}
				if date.Month() == time.Month(monthNum) {
					monthMatch = true
					break
				}
			}

			if date.After(now) && monthDayMatch && monthMatch {
				break
			}

		}
		return date.Format(TmFormat), nil
	}
	return "", nil
}

// функция проверки поля с датой и необходимости ее изменения
func CheckDate(task *db.Task) error {
	now := time.Now()
	//если дата пустая, кидаем текущую
	if task.Date == "" {
		task.Date = now.Format(TmFormat)
		return nil
	}
	//преобразуем в тайм
	t, err := time.Parse(TmFormat, task.Date)
	if err != nil {
		return err
	}
	// если сейчас время больше, чем то, что в задаче, то актуализируем его
	if now.After(t) && (task.Date != now.Format(TmFormat)) {
		if task.Repeat == "" {
			// правило повторения пустое - пишем текущую дату
			task.Date = now.Format(TmFormat)
		} else {
			// правило есть - используем функцию расчета новой даты
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
		}
	}

	return nil
}
