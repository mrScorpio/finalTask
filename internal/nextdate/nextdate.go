package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	date, err := time.Parse("20060102", dstart)
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
	}

	if rep[0] == "d" {
		interval, err := strconv.Atoi(rep[1])
		if err != nil {
			return "", err
		}
		for {
			date = date.AddDate(0, 0, interval)
			if date.After(now) {
				break
			}
		}
	}

	return date.Format("20060102"), nil
}
