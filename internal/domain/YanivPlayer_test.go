package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYanivCardValue(t *testing.T) {
	assert.Equal(t, 0, yanivCardValue(NewCard(CardDesignJoker, 1, false)))
	assert.Equal(t, 0, yanivCardValue(nil))
	assert.Equal(t, 1, yanivCardValue(NewCard(CardDesignSpade, 1, false)))     // Ace
	assert.Equal(t, 7, yanivCardValue(NewCard(CardDesignHeart, 7, false)))     // 7
	assert.Equal(t, 10, yanivCardValue(NewCard(CardDesignClover, 10, false)))  // 10
	assert.Equal(t, 10, yanivCardValue(NewCard(CardDesignDiamond, 13, false))) // King
}

func TestYanivPlayer_HandTotal(t *testing.T) {
	p := NewYanivPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))   // 1
	p.AddCard(NewCard(CardDesignHeart, 13, false))  // 10
	p.AddCard(NewCard(CardDesignJoker, 1, false))   // 0
	p.AddCard(NewCard(CardDesignDiamond, 4, false)) // 4
	assert.Equal(t, 15, p.HandTotal())
}

func TestYanivPlayer_ScoreAndElimination(t *testing.T) {
	p := NewYanivPlayer(false)
	assert.Equal(t, 0, p.GetScore())
	assert.False(t, p.IsEliminated())
	p.AddScore(30)
	p.AddScore(10)
	assert.Equal(t, 40, p.GetScore())
	p.SetScore(5)
	assert.Equal(t, 5, p.GetScore())
	p.SetEliminated(true)
	assert.True(t, p.IsEliminated())
}

func TestYanivPlayer_JSONRoundTrip(t *testing.T) {
	p := NewYanivPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.SetScore(42)
	p.SetEliminated(true)
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored YanivPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 42, restored.GetScore())
	assert.True(t, restored.IsEliminated())
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestYanivPlayer_UnmarshalNilGamePlayer(t *testing.T) {
	var p YanivPlayer
	require.NoError(t, json.Unmarshal([]byte(`{"sc":3}`), &p))
	assert.Equal(t, 3, p.GetScore())
	assert.False(t, p.GetIsHuman())
}
