package be

import "testing"

func Test_isValidTime(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result bool
	}{
		{
			name:   "normal 1",
			input:  "01:01",
			result: true,
		},
		{
			name:   "normal 2",
			input:  "12:59",
			result: true,
		},
		{
			name:   "normal 2",
			input:  "23:59",
			result: true,
		},
		{
			name:   "malformed 1",
			input:  "0959",
			result: false,
		},
		{
			name:   "malformed 2",
			input:  "09 59",
			result: false,
		},
		{
			name:   "malformed 3",
			input:  "a9:59",
			result: false,
		},
		{
			name:   "invalid 1",
			input:  "01:60",
			result: false,
		},
		{
			name:   "invalid 2",
			input:  "25:00",
			result: false,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := isValidTime(testcase.input)
			if result != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
			}
		})
	}
}
