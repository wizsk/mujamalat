package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

func TestStuff2(t *testing.T) {
	d := []HiWord{
		{Idx: 0, Word: "لمنزلي", Future: 1766452427, Past: 1766366027, DontShow: false},
		{Idx: 1, Word: "كنت", Future: 0, Past: 1766366444, DontShow: true},
		{Idx: 2, Word: "فدخلت ", Future: 0, Past: 0, DontShow: false},
		{Idx: 3, Word: "بجفنيه", Future: 0, Past: 0, DontShow: false},
		{Idx: 4, Word: "مكاني ", Future: 0, Past: 0, DontShow: false},
		{Idx: 5, Word: "هذا   ", Future: 0, Past: 0, DontShow: false},
	}

	hm := ordmap.New[string, HiWord]()
	for _, v := range d {
		hm.Set(v.Word, v)
	}

	curr := time.Now().Unix()
	hw, found := hm.GetMatchOrRand(
		func(e *HiWord) bool {
			return !e.DontShow && e.Future < curr
		},
		func(e *HiWord) bool {
			return e.Past == 0 && e.Future == 0
		},
		func(e *HiWord) bool {
			return !e.DontShow
		},
	)

	fmt.Println(found)
	fmt.Println(hw)
}
