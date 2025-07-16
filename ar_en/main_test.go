package ar_en

import "testing"

var at = []struct{ a, t string }{
	{a: "عمل", t: "Eml"},
	{a: "بحه", t: "bHh"},
	{a: "ولم", t: "wlm"},
	{a: "منيستب شسمنيبتسشمبي شمسنيتبمنشسب سمنيبخصثتخقهصتثقىبيرﻻرﻻرشرسشىةؤ", t: "mnystb $smnybts$mby $msnytbmn$sb smnybxSvtxqhStvqYbyrﻻrﻻr$rs$Yp&"},
}

func TestTransliterate(ts *testing.T) {
	for _, c := range at {
		t := transliterate(c.a)
		if t != c.t {
			ts.Fatalf("%s != %s", c.t, t)
		}
	}
}

func TestDeTransliterate(ts *testing.T) {
	for _, c := range at {
		a := deTransliterate(c.t)
		if a != c.a {
			ts.Fatalf("%s != %s", c.a, a)
		}
	}
}
