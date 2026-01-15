package ordmap

import (
	"strconv"
	"strings"
	"testing"
)

func TestGetMatchOrRand(t *testing.T) {
	m := New[int, string]()
	for i := range 5 {
		m.Set(i, "yo")
	}

	// for i := range 5 {
	// 	m.Set(i+8, "box")
	// }

	for i := range 1 {
		m.Set(i+16, "bo")
	}

	t.Log(m.GetMatchOrRand(
		func(s *string) bool { return false },
		func(s *string) bool { return *s == "bo" },
		func(s *string) bool { return *s == "bo" }))

	t.Log(m.GetRand(
		func(s *string) bool { return *s == "bo" }))
}

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
