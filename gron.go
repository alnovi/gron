package gron

import (
	"time"
)

func NextTime(expression string) (time.Time, error) {
	return NextAfter(time.Now(), expression)
}

func NextAfter(after time.Time, expression string) (time.Time, error) {
	expr, err := NewExpr(expression)
	if err != nil {
		return time.Time{}, err
	}
	return expr.NextTime(after), nil
}
