package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDaifugoRankName(t *testing.T) {
	tests := []struct {
		name     string
		rank     int
		expected string
	}{
		{"rank 1", 1, "大富豪"},
		{"rank 2", 2, "富豪"},
		{"rank 3", 3, "平民"},
		{"rank 4", 4, "大貧民"},
		{"unknown rank", 0, "不明"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, daifugoRankName(tt.rank))
		})
	}
}

func TestDaifugoSuitName(t *testing.T) {
	tests := []struct {
		name     string
		suit     int
		expected string
	}{
		{"spade", domain.CardDesignSpade, "SPADE"},
		{"clover", domain.CardDesignClover, "CLOVER"},
		{"heart", domain.CardDesignHeart, "HEART"},
		{"diamond", domain.CardDesignDiamond, "DIAMOND"},
		{"unknown suit", 999, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, daifugoSuitName(tt.suit))
		})
	}
}
