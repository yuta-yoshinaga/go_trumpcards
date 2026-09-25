//go:build test

package domain

import (
	"reflect"
	"testing"
)

func TestSortedIntKeys(t *testing.T) {
	if got := sortedIntKeys(map[int]string{}); len(got) != 0 {
		t.Fatalf("empty map keys = %v, want empty", got)
	}
	got := sortedIntKeys(map[int]string{8: "eight", -2: "minus", 3: "three"})
	if want := []int{-2, 3, 8}; !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
}
