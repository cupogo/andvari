package sqlutil

import (
	"testing"
)

func TestCleanWildcard(t *testing.T) {
	testcases := []struct {
		in  string
		out string
	}{
		{"name*", "name%"},
		{"%name%", "name"},
		{"--;name", "name"},
		{"****name", "name"},
		{"??name?", "name_"},
	}

	for _, tc := range testcases {
		if out := CleanWildcard(tc.in); out != tc.out {
			t.Errorf("CleanWildcard(%q)=%v, expected %v", tc.in, out, tc.out)
		}
	}
}
