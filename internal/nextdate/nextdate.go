// пакет расчета следующей даты при изменении записи
package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
)

// функция возвращает новую дату для задачи, принимает (текущее время, дата задачи, правило повторения)
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// если правило повторения пустое, ничего не делаем
	if repeat == "" {
		return "", nil
	}
	// преобразуем строку с датой в переменную типа тайм
	date, err := time.Parse(db.TmFormat, dstart)
	if err != nil {
		return "", err
	}
	// разбираем строку с правилом повторения в слайс
	rep := strings.Split(repeat, " ")
	// если только один символ и это не год, то плохое правило
	if rep[0] != "y" && len(rep) < 2 {
		return "", fmt.Errorf("wrong repeat format")
	}
	// повторяем год
	if rep[0] == "y" {
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				break
			}

		}
		return date.Format(db.TmFormat), nil
	}
	// повторяем дни
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
		return date.Format(db.TmFormat), nil
	}
	// повторяем дни недели
	if rep[0] == "w" {
		weekDays := strings.Split(rep[1], ",")

		for {
			// приращаем дату
			date = date.AddDate(0, 0, 1)
			weekDayMatch := false
			for _, v := range weekDays {
				weekDayNum, err := strconv.Atoi(v)
				if err != nil {
					return "", err
				}
				// если левый номер дня
				if weekDayNum > 7 || weekDayNum < 0 {
					return "", fmt.Errorf("wrong weekday number")
				}
				// воскресенье в буржуйский формат
				if weekDayNum == 7 {
					weekDayNum = 0
				}
				// если нашли
				if date.Weekday() == time.Weekday(weekDayNum) {
					weekDayMatch = true
					break
				}
			}
			// все совпало - прекращаем поиск
			if date.After(now) && weekDayMatch {
				break
			}

		}
		return date.Format(db.TmFormat), nil
	}

	// повторяем дни месяца
	if rep[0] == "m" {
		monthDays := strings.Split(rep[1], ",")
		// по-умолчанию месяц правильный
		monthMatch := true
		monthNums := make([]string, 0, 12)

		if len(rep) > 2 {
			monthNums = strings.Split(rep[2], ",")
			// но если есть задание месяца в правиле, то нужно проверить
			monthMatch = false
		}
		for {
			// приращаем дату
			date = date.AddDate(0, 0, 1)
			monthDayMatch := false
			prev := false
			postPrev := false
			// цикл для проверки дня
			for _, v := range monthDays {
				monthDay, err := strconv.Atoi(v)
				if err != nil {
					return "", err
				}
				// если левое число
				if monthDay > 31 || monthDay < -2 || monthDay == 0 {
					return "", fmt.Errorf("monthday number is bad")
				}
				// если предпоследнее число
				if monthDay == -2 && date.AddDate(0, 0, 2).Day() == 1 {
					postPrev = true
				}
				// если последнее
				if monthDay == -1 && date.AddDate(0, 0, 1).Day() == 1 {
					prev = true
				}
				// если нашли день
				if date.Day() == monthDay {
					monthDayMatch = true
					break
				}
			}

			if !monthMatch {
				// цикл для проверки месяца
				for _, v := range monthNums {
					monthNum, err := strconv.Atoi(v)
					if err != nil {
						return "", err
					}
					// если левый месяц
					if monthNum > 12 || monthNum < 1 {
						return "", fmt.Errorf("month num is wrong")
					}
					// если нашли месяц
					if date.Month() == time.Month(monthNum) && date.Year() == now.Year() {
						monthMatch = true
						break
					}
				}
			}
			// все совпало - прекращаем поиск
			if date.After(now) && (monthDayMatch || postPrev || prev) && monthMatch {
				break
			}

		}
		return date.Format(db.TmFormat), nil
	}
	return "", nil
}

// функция проверки поля с датой и необходимости ее изменения
func CheckDate(task *db.Task) error {
	now := time.Now()
	//если дата пустая, кидаем текущую
	if task.Date == "" {
		task.Date = now.Format(db.TmFormat)
		return nil
	}
	//преобразуем в тайм
	t, err := time.Parse(db.TmFormat, task.Date)
	if err != nil {
		return err
	}
	// если сейчас время больше, чем то, что в задаче, то актуализируем его
	if now.After(t) && (task.Date != now.Format(db.TmFormat)) {
		if task.Repeat == "" {
			// правило повторения пустое - пишем текущую дату
			task.Date = now.Format(db.TmFormat)
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
