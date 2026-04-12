//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFiftyOnePlayer(t *testing.T) {
	p := NewFiftyOnePlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetCardsSize())

	p2 := NewFiftyOnePlayer(false)
	assert.False(t, p2.GetIsHuman())
}

func TestFiftyOneCardScore(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{"ace", 1, 11},
		{"two", 2, 2},
		{"ten", 10, 10},
		{"jack", 11, 10},
		{"queen", 12, 10},
		{"king", 13, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCard(CardDesignSpade, tt.value, false)
			assert.Equal(t, tt.want, fiftyOneCardScore(c))
		})
	}
}

func TestFiftyOnePlayer_BestSuitScore(t *testing.T) {
	p := NewFiftyOnePlayer(true)
	// 手札なしは0
	assert.Equal(t, 0, p.BestSuitScore())

	// スペードA(11) + スペード10(10) + ハート5(5) = スペード21
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	assert.Equal(t, 21, p.BestSuitScore())

	// ハートK(10)追加 → ハート15 vs スペード21 → 21
	p.AddCard(NewCard(CardDesignHeart, 13, false))
	assert.Equal(t, 21, p.BestSuitScore())

	// スペードK(10)追加 → スペード31 vs ハート15 → 31
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	assert.Equal(t, 31, p.BestSuitScore())
}

func TestFiftyOnePlayer_BestSuitScore_Perfect(t *testing.T) {
	p := NewFiftyOnePlayer(true)
	// 完全ハンド: A(11) + K(10) + Q(10) + J(10) + 10(10) = 51
	p.AddCard(NewCard(CardDesignHeart, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 13, false))
	p.AddCard(NewCard(CardDesignHeart, 12, false))
	p.AddCard(NewCard(CardDesignHeart, 11, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	assert.Equal(t, FiftyOneMaxScore, p.BestSuitScore())
}

func TestFiftyOnePlayer_SuitScores(t *testing.T) {
	p := NewFiftyOnePlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))   // 11
	p.AddCard(NewCard(CardDesignSpade, 5, false))   // 5
	p.AddCard(NewCard(CardDesignHeart, 10, false))  // 10
	p.AddCard(NewCard(CardDesignHeart, 13, false))  // 10
	p.AddCard(NewCard(CardDesignDiamond, 3, false)) // 3

	scores := p.SuitScores()
	assert.Equal(t, 16, scores[CardDesignSpade])
	assert.Equal(t, 20, scores[CardDesignHeart])
	assert.Equal(t, 3, scores[CardDesignDiamond])
	assert.Equal(t, 0, scores[CardDesignClover])
}

func TestFiftyOnePlayer_BestSuit(t *testing.T) {
	p := NewFiftyOnePlayer(true)
	p.AddCard(NewCard(CardDesignHeart, 1, false))  // 11
	p.AddCard(NewCard(CardDesignHeart, 10, false)) // 10
	p.AddCard(NewCard(CardDesignSpade, 5, false))  // 5
	assert.Equal(t, CardDesignHeart, p.BestSuit())
}

func TestFiftyOnePlayer_JSON(t *testing.T) {
	src := NewFiftyOnePlayer(true)
	src.AddCard(NewCard(CardDesignSpade, 1, false))
	src.AddCard(NewCard(CardDesignHeart, 10, false))

	data, err := json.Marshal(src)
	assert.NoError(t, err)

	dst := NewFiftyOnePlayer(false)
	assert.NoError(t, json.Unmarshal(data, dst))
	assert.True(t, dst.GetIsHuman())
	assert.Equal(t, 2, dst.GetCardsSize())
	assert.Equal(t, CardDesignSpade, dst.GetCard(0).GetDesign())
	assert.Equal(t, 1, dst.GetCard(0).GetValue())
}
