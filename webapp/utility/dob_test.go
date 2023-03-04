package utility_test

import (
	"testing"
	"time"

	"github.com/aichaos/silhouette/webapp/utility"
)

func TestAge(t *testing.T) {
	var now = time.Date(2022, 6, 15, 12, 0, 0, 0, time.UTC)

	var tests = []struct {
		In     string
		Expect int
	}{
		{
			In:     "1987-10-12",
			Expect: 34,
		},
		{
			In:     "2022-06-15",
			Expect: 0,
		},
		{
			In:     "1993-07-04",
			Expect: 28,
		},
		{
			In:     "1996-06-17",
			Expect: 26,
		},
		{
			In:     "1996-06-15",
			Expect: 26,
		},
		{
			In:     "1996-06-14",
			Expect: 25,
		},
		{
			In:     "2000-01-01",
			Expect: 22,
		},
		{
			In:     "2000-05-30",
			Expect: 22,
		},
		{
			In:     "2000-06-12",
			Expect: 21,
		},
		{
			In:     "2000-06-14",
			Expect: 21,
		},
		{
			In:     "2000-06-15",
			Expect: 22,
		},
		{
			In:     "2000-06-16",
			Expect: 22,
		},
	}

	for _, test := range tests {
		dob, _ := time.Parse("2006-01-02", test.In)
		actual := utility.AgeAt(dob, now)
		if actual != test.Expect {
			t.Errorf("Expected %s to be age %d but got %d", test.In, test.Expect, actual)
		}
	}
}
