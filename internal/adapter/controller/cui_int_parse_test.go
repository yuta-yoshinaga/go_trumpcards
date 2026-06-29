package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseIntList covers every branch of the shared CUI int-list parser:
// empty input, comma splitting, whitespace trimming, empty-token skipping and
// non-numeric-token skipping.
func TestParseIntList(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []int
	}{
		{name: "nil input returns nil", args: nil, want: nil},
		{name: "empty slice returns nil", args: []string{}, want: nil},
		{name: "single numeric arg", args: []string{"3"}, want: []int{3}},
		{name: "multiple separate args", args: []string{"1", "2", "3"}, want: []int{1, 2, 3}},
		{name: "comma-separated single arg", args: []string{"1,2,3"}, want: []int{1, 2, 3}},
		{name: "whitespace is trimmed", args: []string{" 1 , 2 "}, want: []int{1, 2}},
		{name: "empty tokens are skipped", args: []string{"1,,2"}, want: []int{1, 2}},
		{name: "non-numeric tokens are skipped", args: []string{"1,x,2"}, want: []int{1, 2}},
		{name: "all non-numeric yields nil", args: []string{"a,b"}, want: nil},
		{name: "negative numbers parse", args: []string{"-1,2"}, want: []int{-1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseIntList(tt.args))
		})
	}
}
