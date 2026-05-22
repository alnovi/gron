package gron

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
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
				candidate = candidate.AddDate(0, 0, 1).Truncate(24 * time.Hour) // nolint:mnd
				continue
			}
			candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), nextHour, 0, 0, 0, candidate.Location())
		}

		// Теперь ищем подходящую минуту в этом часе
		if !contains(x.minutes, candidate.Minute()) {
			nextMinute, ok := findNext(partMinutes, x.minutes, candidate.Minute())
			if !ok {
				// Переходим к следующему часу
				candidate = candidate.Add(time.Hour).Truncate(time.Hour)
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

func parseMinutes(val string) ([]int, error) {
	data, err := parseValue(val, partMinutes, 0, 59) // nolint:mnd
	if err != nil {
		return nil, err
	}
	return sortAndUnique(data), nil
}

func parseHours(val string) ([]int, error) {
	data, err := parseValue(val, partHours, 0, 23) // nolint:mnd
	if err != nil {
		return nil, err
	}
	return sortAndUnique(data), nil
}

func parseDaysOfMonth(val string) ([]int, error) {
	data, err := parseValue(val, partDaysOfMonth, 1, 31) // nolint:mnd
	if err != nil {
		return nil, err
	}
	return sortAndUnique(data), nil
}

func parseMonths(val string) ([]int, error) {
	data, err := parseValue(val, partMonth, 1, 12) // nolint:mnd
	if err != nil {
		return nil, err
	}
	return sortAndUnique(data), nil
}

func parseDaysOfWeek(val string) ([]int, error) {
	data, err := parseValue(val, partDaysOfWeek, 0, 6) // nolint:mnd
	if err != nil {
		return nil, err
	}
	return sortAndUnique(data), nil
}

func parseValue(val, part string, start, end int) ([]int, error) {
	data := make([]int, 0, start+1)

	// "*"
	if val == "*" || val == "*/1" {
		return data, nil
	}

	// "L"
	if d, ok := parseValueL(val, part); ok {
		data = append(data, d)
		return data, nil
	}

	// "W"
	if d, ok := parseValueW(val, part); ok {
		data = append(data, d)
		return data, nil
	}

	// "1,3,5", "1,W,L"
	if vals := strings.Split(val, ","); len(vals) > 1 {
		for _, s := range vals {
			if d, ok := parseValueL(s, part); ok {
				data = append(data, d)
				continue
			}

			if d, ok := parseValueW(s, part); ok {
				data = append(data, d)
				continue
			}

			d, err := strconv.Atoi(s)
			if err != nil {
				return []int{}, fmt.Errorf("incorrect %s char: %s", part, s)
			}
			if d < start || d > end {
				return []int{}, fmt.Errorf("%s must be in range [0,6]: %d", part, d)
			}

			data = append(data, d)
		}
		return data, nil
	}

	// "1-5"
	if vals := strings.Split(val, "-"); len(vals) == 2 { // nolint:mnd
		from, err := strconv.Atoi(vals[0])
		if err != nil {
			return []int{}, fmt.Errorf("incorrect %s char: %s", part, vals[0])
		}

		if from < start || from > end {
			return []int{}, fmt.Errorf("%s must be in range [%d,%d]: %d", part, start, end, from)
		}

		to, err := strconv.Atoi(vals[1])
		if err != nil {
			return []int{}, fmt.Errorf("incorrect %s char: %s", part, vals[1])
		}

		if to < start || to > end {
			return []int{}, fmt.Errorf("%s must be in range [%d,%d]: %d", part, start, end, to)
		}

		if from >= to {
			return []int{}, fmt.Errorf("%s start should be greater than the end: [%d-%d]", part, from, to)
		}

		for d := from; d <= to; d++ {
			data = append(data, d)
		}

		return data, nil
	}

	// "*/15"
	if strings.HasPrefix(val, "*/") {
		step, err := strconv.Atoi(strings.Trim(val, "*/"))
		if err != nil || step <= 0 {
			return []int{}, fmt.Errorf("incorrect %s step: %s", part, val)
		}

		for d := start; d <= end; d += step {
			data = append(data, d)
		}

		return data, nil
	}

	// "1"
	if d, err := strconv.Atoi(val); err == nil {
		if d < start || d > end {
			return []int{}, fmt.Errorf("%s must be in range [%d,%d]: %d", part, start, end, d)
		}
		data = append(data, d)
		return data, nil
	}

	return nil, fmt.Errorf("invalid %s: %s", part, val)
}

func parseValueL(val, part string) (int, bool) {
	return L, part == partDaysOfMonth && val == "L"
}

func parseValueW(val, part string) (int, bool) {
	return W, part == partDaysOfMonth && val == "W"
}

func sortAndUnique(data []int) []int {
	seen := make(map[int]bool)
	var unique []int

	for _, num := range data {
		if !seen[num] {
			seen[num] = true
			unique = append(unique, num)
		}
	}

	slices.Sort(unique)
	return unique
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
