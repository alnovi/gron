package gron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExprSuccessExpr(t *testing.T) {
	testCases := []struct {
		expression string
		expected   *Expr
	}{
		{
			expression: "* * * * *",
			expected: &Expr{
				minutes:     nil,
				hours:       nil,
				daysOfMonth: nil,
				months:      nil,
				daysOfWeek:  nil,
			},
		},
		{
			expression: "0 * L * *",
			expected: &Expr{
				minutes:     []int{0},
				hours:       nil,
				daysOfMonth: []int{L},
				months:      nil,
				daysOfWeek:  nil,
			},
		},
		{
			expression: "1 * W * *",
			expected: &Expr{
				minutes:     []int{1},
				hours:       nil,
				daysOfMonth: []int{W},
				months:      nil,
				daysOfWeek:  nil,
			},
		},
		{
			expression: "1 1 1 1 1",
			expected: &Expr{
				minutes:     []int{1},
				hours:       []int{1},
				daysOfMonth: []int{1},
				months:      []int{1},
				daysOfWeek:  []int{1},
			},
		},
		{
			expression: "1 1 1 * *",
			expected: &Expr{
				minutes:     []int{1},
				hours:       []int{1},
				daysOfMonth: []int{1},
				months:      nil,
				daysOfWeek:  nil,
			},
		},
		{
			expression: "59 12 31 12 6",
			expected: &Expr{
				minutes:     []int{59},
				hours:       []int{12},
				daysOfMonth: []int{31},
				months:      []int{12},
				daysOfWeek:  []int{6},
			},
		},
		{
			expression: "*/10 */2 */10 */3 */2",
			expected: &Expr{
				minutes:     []int{0, 10, 20, 30, 40, 50},
				hours:       []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22},
				daysOfMonth: []int{1, 11, 21, 31},
				months:      []int{1, 4, 7, 10},
				daysOfWeek:  []int{0, 2, 4, 6},
			},
		},
		{
			expression: "0-5 0-5 1-5 1-5 1-3",
			expected: &Expr{
				minutes:     []int{0, 1, 2, 3, 4, 5},
				hours:       []int{0, 1, 2, 3, 4, 5},
				daysOfMonth: []int{1, 2, 3, 4, 5},
				months:      []int{1, 2, 3, 4, 5},
				daysOfWeek:  []int{1, 2, 3},
			},
		},
		{
			expression: "0,1,2,2 0,1,2,0 1,2,3,1 1,2,3,2 1,2,3,3",
			expected: &Expr{
				minutes:     []int{0, 1, 2},
				hours:       []int{0, 1, 2},
				daysOfMonth: []int{1, 2, 3},
				months:      []int{1, 2, 3},
				daysOfWeek:  []int{1, 2, 3},
			},
		},
		{
			expression: "* * 1,L,W * *",
			expected: &Expr{
				minutes:     nil,
				hours:       nil,
				daysOfMonth: []int{W, L, 1},
				months:      nil,
				daysOfWeek:  nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			actual, err := NewExpr(tc.expression)
			require.NoError(t, err, "failed to parse expression")

			assert.Equal(t, tc.expected.minutes, actual.minutes, "invalid minutes")
			assert.Equal(t, tc.expected.hours, actual.hours, "invalid hours")
			assert.Equal(t, tc.expected.daysOfMonth, actual.daysOfMonth, "invalid days of month")
			assert.Equal(t, tc.expected.months, actual.months, "invalid months")
			assert.Equal(t, tc.expected.daysOfWeek, actual.daysOfWeek, "invalid days of week")
		})
	}
}

func TestExprFailureExpr(t *testing.T) {
	testCases := []struct {
		expression string
		expected   string
	}{
		{
			expression: "* * *",
			expected:   "invalid cron expression",
		},
		{
			expression: "60 * * * *",
			expected:   "minutes must be in range [0,59]",
		},
		{
			expression: "* 24 * * *",
			expected:   "hours must be in range [0,23]",
		},
		{
			expression: "* * 0 * *",
			expected:   "days of month must be in range [1,31]",
		},
		{
			expression: "* * * 0 *",
			expected:   "months must be in range [1,12]",
		},
		{
			expression: "* * * * 7",
			expected:   "days of week must be in range [0,6]",
		},
		{
			expression: "a-5 * * * *",
			expected:   "incorrect minutes char: a",
		},
		{
			expression: "1-a * * * *",
			expected:   "incorrect minutes char: a",
		},
		{
			expression: "10-5 * * * *",
			expected:   "minutes start should be greater than the end: [10-5]",
		},
		{
			expression: "60-61 * * * *",
			expected:   "minutes must be in range [0,59]",
		},
		{
			expression: "10-61 * * * *",
			expected:   "minutes must be in range [0,59]",
		},
		{
			expression: "* * 1,A * *",
			expected:   "incorrect days of month char: A",
		},
		{
			expression: "* * 0,1 * *",
			expected:   "days of month must be in range [0,6]",
		},
		{
			expression: "* * */0 * *",
			expected:   "incorrect days of month step: */0",
		},
		{
			expression: "* * 5#3 * *",
			expected:   "invalid days of month: 5#3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			_, err := NewExpr(tc.expression)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expected)
		})
	}
}

func TestExprMatches(t *testing.T) {
	testCases := []struct {
		expression string
		date       time.Time
		expected   bool
	}{
		{
			expression: "10 * * * *",
			date:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   false,
		},
		{
			expression: "* 10 * * *",
			date:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   false,
		},
		{
			expression: "* * * 10 *",
			date:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			expr, err := NewExpr(tc.expression)
			require.NoError(t, err)

			actual := expr.Matches(tc.date)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
