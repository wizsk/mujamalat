package ordmap

import (
	"strconv"
	"strings"
	"testing"
)

func TestJoinStr(t *testing.T) {
	s := New[string, struct{}]()
	var a []string

	// emtpy
	if strings.Join(a, "\n") != s.JoinStr(func(e Entry[string, struct{}]) string { return e.Key }, "\n") {
		t.Log("NO match empty")
		t.FailNow()
	}

	for i := range 10_000 {
		l := strconv.Itoa(i)
		a = append(a, l)
		s.Set(l, struct{}{})
	}

	if strings.Join(a, "\n") != s.JoinStr(func(e Entry[string, struct{}]) string { return e.Key }, "\n") {
		t.Log("NO match")
		t.FailNow()
	}
}
