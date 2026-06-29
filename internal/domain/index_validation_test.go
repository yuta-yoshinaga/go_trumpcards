//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateIndexList covers every branch of the shared index-list validator:
// empty list, out-of-range (negative and too-large), duplicates and the valid
// single / multiple cases.
func TestValidateIndexList(t *testing.T) {
	tests := []struct {
		name    string
		indices []int
		size    int
		wantErr bool
	}{
		{name: "empty list is invalid", indices: []int{}, size: 5, wantErr: true},
		{name: "nil list is invalid", indices: nil, size: 5, wantErr: true},
		{name: "negative index is out of range", indices: []int{-1}, size: 5, wantErr: true},
		{name: "index equal to size is out of range", indices: []int{5}, size: 5, wantErr: true},
		{name: "index beyond size is out of range", indices: []int{6}, size: 5, wantErr: true},
		{name: "duplicate index is invalid", indices: []int{1, 1}, size: 5, wantErr: true},
		{name: "single valid index", indices: []int{0}, size: 5, wantErr: false},
		{name: "multiple distinct valid indices", indices: []int{0, 2, 4}, size: 5, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIndexList(tt.indices, tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
