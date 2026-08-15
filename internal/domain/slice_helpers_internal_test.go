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

// filterIndices replaced 26 byte-identical per-game copies (koenigrufenFilter,
// gzFilter, twentyNineFilter, …). They differed only in name, which is why a
// name-based search never grouped them — see issue #5361.
func TestFilterIndices(t *testing.T) {
	tests := []struct {
		name    string
		indices []int
		pred    func(int) bool
		want    []int
	}{
		{"keeps the matching indices", []int{1, 2, 3, 4}, func(i int) bool { return i%2 == 0 }, []int{2, 4}},
		{"keeps everything when the predicate always holds", []int{7, 8}, func(int) bool { return true }, []int{7, 8}},
		// nil, not an empty slice: the copies all started from `var out []int`
		// and appended, so a caller that distinguishes the two keeps working.
		{"returns nil when nothing matches", []int{1, 3}, func(i int) bool { return i%2 == 0 }, nil},
		{"returns nil for an empty input", []int{}, func(int) bool { return true }, nil},
		{"returns nil for a nil input", nil, func(int) bool { return true }, nil},
		{"preserves order and duplicates", []int{5, 1, 5, 2}, func(i int) bool { return i != 2 }, []int{5, 1, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIndices(tt.indices, tt.pred)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The input must not be modified: several callers pass a slice they keep using
// (a hand's valid-index list) and then filter it again with a second predicate.
func TestFilterIndices_DoesNotMutateInput(t *testing.T) {
	in := []int{1, 2, 3, 4}
	_ = filterIndices(in, func(i int) bool { return i > 2 })
	if want := []int{1, 2, 3, 4}; !reflect.DeepEqual(in, want) {
		t.Errorf("input mutated: got %v, want %v", in, want)
	}
}
