package main

import "testing"

func TestCleanSpacesInPlace(t *testing.T) {
	cases := []struct{ in, ex string }{
		{
			"      ",
			"",
		},
		{
			" x\n  \n   y",
			"x\ny",
		},
		{
			"\nb \n😐c  \n  d 🤡   \n  ",
			"b\n😐c\nd 🤡",
		},
		{
			"       GH t g. غ \n ",
			"GH t g. غ",
		},
	}
	for i, c := range cases {
		r := string(cleanSpacesInPlace([]byte(c.in)))
		if r != c.ex {
			t.Logf("case: %d:%q", i, c.in)
			t.Logf("ex:%q != got:%q", c.ex, r)
			t.FailNow()
		}
	}
}
