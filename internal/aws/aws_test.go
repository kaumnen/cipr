package aws

import (
	"testing"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestSeparateFilters(t *testing.T) {
	logger := utils.GetCiprLogger()

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
		logger.Printf("Running test case '%s' - input: %s | expected: %v",
			tc.name, tc.input, tc.expected)

		result := separateFilters(tc.input)
		assert.Equal(t, tc.expected, result, "Test case '%s' failed: input=%s, expected=%v, got=%v",
			tc.name, tc.input, tc.expected, result)
	}
}

func TestFiltrateIPRanges(t *testing.T) {
	logger := utils.GetCiprLogger()

	testCases := []struct {
		name        string
		ipType      string
		filterSlice []string
		expected    []string
	}{
		{
			name:        "ipv4 - filters: us-east-1 region, EBS service",
			ipType:      "ipv4",
			filterSlice: []string{"us-east-1", "EBS"},
			expected: []string{
				"44.192.140.112/28,us-east-1,EBS,us-east-1",
				"44.192.140.128/29,us-east-1,EBS,us-east-1",
				"44.222.159.166/31,us-east-1,EBS,us-east-1",
				"44.222.159.176/28,us-east-1,EBS,us-east-1",
			},
		},
		{
			name:        "ipv6 - filters: us-east-1 region, EBS service, us-east-1 network border group",
			ipType:      "ipv6",
			filterSlice: []string{"eu-central-1", "S3", "eu-central-1"},
			expected: []string{
				"2a05:d070:4000::/40,eu-central-1,S3,eu-central-1",
				"2a05:d079:4000::/40,eu-central-1,S3,eu-central-1",
				"2a05:d034:4000::/40,eu-central-1,S3,eu-central-1",
				"2a05:d07a:4000::/40,eu-central-1,S3,eu-central-1",
				"2a05:d078:4000::/40,eu-central-1,S3,eu-central-1",
				"2a05:d050:4000::/40,eu-central-1,S3,eu-central-1",
			},
		},
	}

	for _, tc := range testCases {
		logger.Printf("Running test case '%s' - ip type: %s | filter slice: %v | expected: %v",
			tc.name, tc.ipType, tc.filterSlice, tc.expected)

		rawData := mockGetReq()

		result := filtrateIPRanges(rawData, tc.ipType, tc.filterSlice)
		assert.Equal(t, tc.expected, result, "Test case '%s' failed: expected=%v, got=%v",
			tc.name, tc.expected, result)
	}
}
