package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThirtyOnePlayer_LivesAndElimination(t *testing.T) {
	p := NewThirtyOnePlayer(false)
	p.SetLives(1)
	assert.Equal(t, 1, p.GetLives())
	assert.False(t, p.IsEliminated())
	p.LoseLife()
	assert.Equal(t, 0, p.GetLives())
	assert.False(t, p.IsEliminated())
	p.LoseLife()
	assert.Equal(t, -1, p.GetLives())
	assert.True(t, p.IsEliminated())
}

func TestThirtyOnePlayer_SuitScoresMap(t *testing.T) {
	p := NewThirtyOnePlayer(true)
	p.AddCard(NewCard(CardDesignHeart, 1, false))  // 11
	p.AddCard(NewCard(CardDesignHeart, 11, false)) // 10 (J)
	p.AddCard(NewCard(CardDesignSpade, 7, false))  // 7
	scores := p.SuitScores()
	assert.Equal(t, 21, scores[CardDesignHeart])
	assert.Equal(t, 7, scores[CardDesignSpade])
	assert.Equal(t, 0, scores[CardDesignClover])
	assert.Equal(t, 21, p.BestSuitScore())
	assert.Equal(t, CardDesignHeart, p.BestSuit())
}

func TestThirtyOnePlayer_EmptyHand(t *testing.T) {
	p := NewThirtyOnePlayer(false)
	assert.Equal(t, 0, p.BestSuitScore())
	assert.Equal(t, CardDesignSpade, p.BestSuit())
}

func TestThirtyOnePlayer_JSONRoundTrip(t *testing.T) {
	p := NewThirtyOnePlayer(true)
	p.SetLives(2)
	p.AddCard(NewCard(CardDesignDiamond, 9, false))
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored ThirtyOnePlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 2, restored.GetLives())
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestThirtyOnePlayer_UnmarshalNilGamePlayer(t *testing.T) {
	var p ThirtyOnePlayer
	require.NoError(t, json.Unmarshal([]byte(`{"lv":5}`), &p))
	assert.Equal(t, 5, p.GetLives())
	assert.NotNil(t, p.GamePlayer)
}
