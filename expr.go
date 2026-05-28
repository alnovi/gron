package gron

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	partMinutes     = "minutes"
	partHours       = "hours"
	partDaysOfMonth = "days of month"
	partMonth       = "months"
	partDaysOfWeek  = "days of week"
	L               = -1
	W               = -2
)

var ErrInvalidCronExpression = errors.New("invalid cron expression")

type Expr struct {
	minutes     []int // Минуты 0-59
	hours       []int // Часы 0-23
	daysOfMonth []int // День месяца 1-31, где L=-1 W=-2
	months      []int // Месяц 1-12
	daysOfWeek  []int // День недели 0-6, где 0 — воскресенье
}

func NewExpr(expression string) (*Expr, error) {
	var err error

	expr := &Expr{}

	vals := strings.Split(expression, " ")
	if len(vals) != 5 { // nolint:mnd
		return expr, ErrInvalidCronExpression
	}

	expr.minutes, err = parseMinutes(vals[0])
	if err != nil {
		return expr, fmt.Errorf("%w: %w", ErrInvalidCronExpression, err)
	}

	expr.hours, err = parseHours(vals[1])
	if err != nil {
		return expr, fmt.Errorf("%w: %w", ErrInvalidCronExpression, err)
	}

	expr.daysOfMonth, err = parseDaysOfMonth(vals[2])
	if err != nil {
		return expr, fmt.Errorf("%w: %w", ErrInvalidCronExpression, err)
	}

	expr.months, err = parseMonths(vals[3])
	if err != nil {
		return expr, fmt.Errorf("%w: %w", ErrInvalidCronExpression, err)
	}

	expr.daysOfWeek, err = parseDaysOfWeek(vals[4])
	if err != nil {
		return expr, fmt.Errorf("%w: %w", ErrInvalidCronExpression, err)
	}

	return expr, nil
}

func (x *Expr) NextTime(date time.Time) time.Time {
	candidate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, date.Location())
	candidate = candidate.Add(time.Minute)

	for {
		// Быстро переходим к ближайшему месяцу
		if !contains(x.months, int(candidate.Month())) {
			nextMonth, ok := findNext(partMonth, x.months, int(candidate.Month()))
			if !ok {
				// Переходим на следующий год
				candidate = time.Date(candidate.Year()+1, 1, 0, 0, 0, 0, 0, date.Location())
				continue
			}
			candidate = time.Date(candidate.Year(), time.Month(nextMonth), 1, 0, 0, 0, 0, candidate.Location())
		}

		// Теперь ищем подходящий день
		if !contains(x.daysOfMonth, L) &&
			!contains(x.daysOfMonth, W) &&
			!contains(x.daysOfMonth, candidate.Day()) &&
			len(x.daysOfWeek) == 0 {
			nextDay, ok := findNext(partDaysOfMonth, x.daysOfMonth, candidate.Day())
			if !ok {
				// Переходим на следующий месяц
				candidate = time.Date(candidate.Year(), candidate.Month()+1, 0, 0, 0, 0, 0, date.Location())
				continue
			}
			candidate = time.Date(candidate.Year(), candidate.Month(), nextDay, 0, 0, 0, 0, candidate.Location())
		}

		// Теперь ищем подходящий часу в этом дне
		if !contains(x.hours, candidate.Hour()) {
			nextHour, ok := findNext(partHours, x.hours, candidate.Hour())
			if !ok {
				// Переходим на следующий день
				candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day()+1, 0, 0, 0, 0, candidate.Location())
				continue
			}
			candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), nextHour, 0, 0, 0, candidate.Location())
		}

		// Теперь ищем подходящую минуту в этом часе
		if !contains(x.minutes, candidate.Minute()) {
			nextMinute, ok := findNext(partMinutes, x.minutes, candidate.Minute())
			if !ok {
				// Переходим к следующему часу
				candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), candidate.Hour()+1, 0, 0, 0, candidate.Location())
				continue
			}
			candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), candidate.Hour(), nextMinute, 0, 0, candidate.Location())
		}

		if x.Matches(candidate) {
			return candidate
		}

		candidate = candidate.Add(time.Minute)
	}
}

func (x *Expr) Matches(date time.Time) bool {
	if !contains(x.minutes, date.Minute()) {
		return false
	}

	if !contains(x.hours, date.Hour()) {
		return false
	}

	if !contains(x.months, int(date.Month())) {
		return false
	}

	// Обрабатываем день месяца и день недели (cron использует OR)
	dayOfMonthMatch := matchDayOfMonth(x.daysOfMonth, date)
	dayOfWeekMatch := contains(x.daysOfWeek, int(date.Weekday()))

	// Если оба поля не пустые, достаточно выполнения одного условия
	if len(x.daysOfMonth) > 0 && len(x.daysOfWeek) > 0 {
		return dayOfMonthMatch || dayOfWeekMatch
	}

	// Иначе проверяем только заполненное поле
	if len(x.daysOfMonth) > 0 {
		return dayOfMonthMatch
	}
	return dayOfWeekMatch
}

func findNext(part string, values []int, current int) (int, bool) {
	for _, v := range values {
		if v > current {
			return v, true
		}
	}

	if len(values) == 0 {
		for _, v := range defaultValues(part) {
			if v > current {
				return v, true
			}
		}
	}

	return 0, false
}

func defaultValues(part string) []int {
	data := make([]int, 0)

	switch part {
	case partMinutes:
		for m := range 60 {
			data = append(data, m)
		}
	case partHours:
		for h := range 24 {
			data = append(data, h)
		}
	case partDaysOfMonth:
		for d := range 31 {
			data = append(data, d+1)
		}
	case partMonth:
		for m := range 12 {
			data = append(data, m+1)
		}
	case partDaysOfWeek:
		for d := range 7 {
			data = append(data, d)
		}
	}

	return data
}

func contains(s []int, v int) bool {
	if len(s) == 0 {
		return true
	}
	return slices.Contains(s, v)
}

func matchDayOfMonth(days []int, date time.Time) bool {
	if slices.Contains(days, L) {
		if date.Month() != date.AddDate(0, 0, 1).Month() {
			return true
		}
	}

	if slices.Contains(days, W) {
		if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday {
			return true
		}
	}

	return contains(days, date.Day())
}
