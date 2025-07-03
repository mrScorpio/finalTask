package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mrScorpio/finalTask/internal/db"
)

const TmFormat string = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	date, err := time.Parse(TmFormat, dstart)
	if err != nil {
		return "", err
	}

	rep := strings.Split(repeat, " ")

	if rep[0] != "y" && len(rep) < 2 {
		return "", fmt.Errorf("wrong repeat format")
	}

	if rep[0] == "y" {
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				break
			}

		}
		return date.Format(TmFormat), nil
	}

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

	return "", nil
}

func CheckDate(task *db.Task) error {
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(TmFormat)
		return nil
	}

	t, err := time.Parse(TmFormat, task.Date)
	if err != nil {
		return err
	}

	if now.After(t) && (task.Date != now.Format(TmFormat)) {
		if task.Repeat == "" {
			task.Date = now.Format(TmFormat)
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
		}
	}

	return nil
}
