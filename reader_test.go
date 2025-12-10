package main

import (
	"fmt"
	"slices"
	"testing"
)

func TestStuff(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4, 5}
	sIdxs := []int{0, 2, 3}
	res := []int{1, 4, 5}

	i := 0
	curr := sIdxs[i]

	for next := sIdxs[i]; next < len(arr); {
		if len(sIdxs) > i && next == sIdxs[i] {
			fmt.Println("skipping", next)
			i++
			next++
			continue
		}
		fmt.Println("adding", curr, next)
		arr[curr] = arr[next]
		curr++
		next++
	}
	arr = arr[:curr]

	if !slices.Equal(arr, res) {
		t.Log("ex", res)
		t.Log("out", arr)
		t.FailNow()
	}
}
