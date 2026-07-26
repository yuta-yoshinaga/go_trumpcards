//go:build test

package domain

import (
	"reflect"
	"testing"
)

func TestCollectValidIndices(t *testing.T) {
	tests := []struct {
		name string
		size int
		ok   func(i int) bool
		want []int
	}{
		{"all valid", 3, func(int) bool { return true }, []int{0, 1, 2}},
		{"none valid", 3, func(int) bool { return false }, []int{}},
		{"even only", 5, func(i int) bool { return i%2 == 0 }, []int{0, 2, 4}},
		{"empty hand", 0, func(int) bool { return true }, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := collectValidIndices(tc.size, tc.ok)
			if got == nil {
				t.Fatalf("result must be non-nil even when empty")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("collectValidIndices(%d) = %v, want %v", tc.size, got, tc.want)
			}
		})
	}
}
