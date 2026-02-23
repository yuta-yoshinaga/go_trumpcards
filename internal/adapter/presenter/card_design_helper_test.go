package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCardDesignToString(t *testing.T) {
	tests := []struct {
		name     string
		design   int
		expected string
	}{
		{"spade", domain.CardDesignSpade, "SPADE"},
		{"clover", domain.CardDesignClover, "CLOVER"},
		{"heart", domain.CardDesignHeart, "HEART"},
		{"diamond", domain.CardDesignDiamond, "DIAMOND"},
		{"joker", domain.CardDesignJoker, "JOKER"},
		{"unknown design", 999, "JOKER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cardDesignToString(tt.design))
		})
	}
}
