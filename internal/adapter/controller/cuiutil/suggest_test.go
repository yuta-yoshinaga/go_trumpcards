//go:build test

package cuiutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"empty strings", "", "", 0},
		{"first empty", "", "abc", 3},
		{"second empty", "abc", "", 3},
		{"identical", "hit", "hit", 0},
		{"single substitution", "hit", "hut", 1},
		{"transposition", "hti", "hit", 2},
		{"missing char", "ht", "hit", 1},
		{"extra char", "hiit", "hit", 1},
		{"completely different", "abc", "xyz", 3},
		{"case sensitive", "Hit", "hit", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, LevenshteinDistance(tt.a, tt.b))
		})
	}
}

func TestSuggestCommand(t *testing.T) {
	commands := []string{"hit", "stand", "bet", "doubledown", "split"}

	tests := []struct {
		name     string
		input    string
		commands []string
		maxDist  int
		want     string
	}{
		{"exact match", "hit", commands, 3, "hit"},
		{"typo transposition", "hti", commands, 3, "hit"},
		{"missing char", "ht", commands, 3, "hit"},
		{"extra char", "hiit", commands, 3, "hit"},
		{"substitution", "hut", commands, 3, "hit"},
		{"stand typo", "stnad", commands, 3, "stand"},
		{"bet typo", "bte", commands, 3, "bet"},
		{"too far away", "zzzzz", commands, 3, ""},
		{"empty input", "", commands, 3, ""},
		{"empty commands", "hit", nil, 3, ""},
		{"zero max distance", "hti", commands, 0, ""},
		{"picks closest match", "splitt", commands, 3, "split"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SuggestCommand(tt.input, tt.commands, tt.maxDist))
		})
	}
}
