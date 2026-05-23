package gron

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

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
	if d, ok := parseMonthLastDay(val, part); ok {
		data = append(data, d)
		return data, nil
	}

	// "W"
	if d, ok := parseMonthWorkingDay(val, part); ok {
		data = append(data, d)
		return data, nil
	}

	// "1,3,5", "1,W,L"
	if vals := strings.Split(val, ","); len(vals) > 1 {
		for _, s := range vals {
			if d, ok := parseMonthLastDay(s, part); ok {
				data = append(data, d)
				continue
			}

			if d, ok := parseMonthWorkingDay(s, part); ok {
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

func parseMonthLastDay(val, part string) (int, bool) {
	return L, part == partDaysOfMonth && val == "L"
}

func parseMonthWorkingDay(val, part string) (int, bool) {
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
