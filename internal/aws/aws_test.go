package aws

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeparateFilters(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Three items",
			input:    "one,two,three",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Two items",
			input:    "one,two",
			expected: []string{"one", "two"},
		},
		{
			name:     "Single item",
			input:    "single,",
			expected: []string{"single"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nRunning test case '%s' - input: %s | expected: %v\n\n----------------------\n",
			tc.name, tc.input, tc.expected)

		result := separateFilters(tc.input)
		assert.Equal(t, tc.expected, result, "Test case '%s' failed: input=%s, expected=%v, got=%v",
			tc.name, tc.input, tc.expected, result)
	}
}
