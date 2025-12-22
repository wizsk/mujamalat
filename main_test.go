package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestDurToDHM(t *testing.T) {
	oneDay := time.Hour * 24
	cases := []struct {
		in time.Duration
		ex string
	}{
		{
			in: oneDay,
			ex: "1d later",
		},
		{
			in: oneDay + time.Hour + time.Minute,
			ex: "1d 1h 1m later",
		},
		{
			in: (oneDay * 30) + time.Hour + time.Minute,
			ex: "30d 1h 1m later",
		},
	}
	for i, c := range cases {
		r := durToDHM(c.in, false)
		if r != c.ex {
			t.Logf("case: %d:%q", i, c.in)
			t.Logf("ex:%q != got:%q", c.ex, r)
			t.FailNow()
		}
	}
}
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

func TestFormatInputText(t *testing.T) {
	cases := []struct{ in, ex []byte }{
		{
			[]byte(""),
			[]byte(""),
		},
		{
			[]byte("  \n \t             "),
			[]byte(""),
		},
		{
			[]byte("  \n \t   #          "),
			[]byte(magicValMJENnl + ":#\n\n"),
		},
		{
			[]byte(`أَلا أَيُّهَا اَلمَقْصُودُ؟ فِي كُلِّ! حَاجَةٍ،
شَكَوْتُ إِلَيْكَ الضُّرَّ - فَارْحَمْ شَكَايَتِي.
٢`),
			[]byte(magicValMJENnl + `أَلا:أَلا
أَيُّهَا:أَيُّهَا
اَلمَقْصُودُ:اَلمَقْصُودُ؟
فِي:فِي
كُلِّ:كُلِّ!
حَاجَةٍ:حَاجَةٍ،

شَكَوْتُ:شَكَوْتُ
إِلَيْكَ:إِلَيْكَ
الضُّرَّ:الضُّرَّ
:-
فَارْحَمْ:فَارْحَمْ
شَكَايَتِي:شَكَايَتِي.

:٢

`),
		},
	}
	buf := new(bytes.Buffer)
	for i, c := range cases {
		formatInputText(c.in, buf)
		if !bytes.Equal(buf.Bytes(), c.ex) {
			fmt.Printf("case: %d:\n%s\n", i, c.in)
			fmt.Printf("  ex: \n%s\n", c.ex)
			fmt.Printf(" got: \n%s\n", buf.Bytes())

			fmt.Printf("  ex: %q\n", c.ex)
			fmt.Printf(" got: %q\n", buf.Bytes())
			t.FailNow()
		}
	}
}
