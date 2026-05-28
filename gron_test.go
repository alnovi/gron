package gron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGronNextTime(t *testing.T) {
	_, err := NextTime("* * *")
	require.Error(t, err)
}

func TestGronNextAfter(t *testing.T) {
	testCases := []struct {
		expression string
		after      time.Time
		expected   time.Time
	}{
		{
			expression: "* * * * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC),
		},
		{
			expression: "10 * * * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC),
		},
		{
			expression: "20 * * * *",
			after:      time.Date(2026, 1, 1, 0, 50, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 1, 20, 0, 0, time.UTC),
		},
		{
			expression: "* 2 * * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC),
		},
		{
			expression: "0 2,3 * * *",
			after:      time.Date(2026, 1, 1, 2, 1, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC),
		},
		{
			expression: "0 0 * * *",
			after:      time.Date(2026, 1, 1, 2, 1, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "0 0 21 * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* * 31 * *",
			after:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "0 0 3 2 *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* * L * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* * 30,L 2 *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* * 10 * 2",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* * 2 * 6",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			expression: "* 23 W * *",
			after:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2026, 1, 1, 23, 0, 0, 0, time.UTC),
		},
		{
			expression: "* 1 31 * *",
			after:      time.Date(2026, 12, 31, 10, 0, 0, 0, time.UTC),
			expected:   time.Date(2027, 1, 31, 1, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expression, func(t *testing.T) {
			actual, err := NextAfter(tc.after, tc.expression)
			require.NoError(t, err, "failed to get next time")
			assert.Equal(t, tc.expected, actual, "invalid next time")
		})
	}
}
