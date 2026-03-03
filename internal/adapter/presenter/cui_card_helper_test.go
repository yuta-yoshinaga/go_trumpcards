package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCuiCardStr(t *testing.T) {
	tests := []struct {
		name     string
		card     *domain.Card
		expected string
	}{
		{"nil card", nil, "??"},
		{"joker", domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false), "JOKER"},
		{"spade", domain.NewCard(domain.CardDesignSpade, 1, false), "SPADE 1"},
		{"clover", domain.NewCard(domain.CardDesignClover, 5, false), "CLOVER 5"},
		{"heart", domain.NewCard(domain.CardDesignHeart, 10, false), "HEART 10"},
		{"diamond", domain.NewCard(domain.CardDesignDiamond, 13, false), "DIAMOND 13"},
		{"unknown design", domain.NewCard(99, 1, false), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardStr(tt.card))
		})
	}
}

func TestCuiCardStrEmoji(t *testing.T) {
	tests := []struct {
		name     string
		card     *domain.Card
		expected string
	}{
		{"nil card", nil, "??"},
		{"joker", domain.NewCard(domain.CardDesignJoker, 0, false), "🃏0"},
		{"spade", domain.NewCard(domain.CardDesignSpade, 1, false), "♠1"},
		{"clover", domain.NewCard(domain.CardDesignClover, 5, false), "♣5"},
		{"heart", domain.NewCard(domain.CardDesignHeart, 10, false), "♥10"},
		{"diamond", domain.NewCard(domain.CardDesignDiamond, 13, false), "♦13"},
		{"negative design falls back to joker", domain.NewCard(-1, 7, false), "🃏7"},
		{"out-of-range design falls back to joker", domain.NewCard(99, 3, false), "🃏3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardStrEmoji(tt.card))
		})
	}
}

func TestCuiSuitName(t *testing.T) {
	tests := []struct {
		name     string
		suit     int
		expected string
	}{
		{"spade", domain.CardDesignSpade, "SPADE"},
		{"clover", domain.CardDesignClover, "CLOVER"},
		{"heart", domain.CardDesignHeart, "HEART"},
		{"diamond", domain.CardDesignDiamond, "DIAMOND"},
		{"unknown", 999, "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiSuitName(tt.suit))
		})
	}
}

// mockCuiPlayer implements the cuiPlayer interface for testing.
type mockCuiPlayer struct {
	isHuman bool
}

func (m *mockCuiPlayer) GetIsHuman() bool {
	return m.isHuman
}

func TestCuiPlayerName(t *testing.T) {
	tests := []struct {
		name     string
		player   *mockCuiPlayer
		idx      int
		expected string
	}{
		{"nil player", nil, 0, "UNKNOWN"},
		{"human player", &mockCuiPlayer{isHuman: true}, 0, "あなた"},
		{"cpu player idx 1", &mockCuiPlayer{isHuman: false}, 1, "CPU 1"},
		{"cpu player idx 3", &mockCuiPlayer{isHuman: false}, 3, "CPU 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiPlayerName(tt.player, tt.idx))
		})
	}
}

func TestCuiBettingActionName(t *testing.T) {
	tests := []struct {
		name     string
		action   int
		expected string
	}{
		{"fold", domain.PokerActionFold, "フォールド"},
		{"check", domain.PokerActionCheck, "チェック"},
		{"call", domain.PokerActionCall, "コール"},
		{"bet", domain.PokerActionBet, "ベット"},
		{"raise", domain.PokerActionRaise, "レイズ"},
		{"all-in", domain.PokerActionAllIn, "オールイン"},
		{"unknown", 999, "不明"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiBettingActionName(tt.action))
		})
	}
}
