//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoFishPlayer_NewGoFishPlayer(t *testing.T) {
	p := NewGoFishPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetBookCount())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestGoFishPlayer_Books(t *testing.T) {
	p := NewGoFishPlayer(false)
	cards := []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
	}
	p.AddBook(cards)
	assert.Equal(t, 1, p.GetBookCount())
	assert.Equal(t, cards, p.GetBooks()[0])

	p.ResetBooks()
	assert.Equal(t, 0, p.GetBookCount())
}

func TestGoFishPlayer_HasRank(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 7, false))

	assert.True(t, p.HasRank(3))
	assert.True(t, p.HasRank(7))
	assert.False(t, p.HasRank(1))
}

func TestGoFishPlayer_CountRank(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 3, false))
	p.AddCard(NewCard(CardDesignClover, 5, false))

	assert.Equal(t, 2, p.CountRank(3))
	assert.Equal(t, 1, p.CountRank(5))
	assert.Equal(t, 0, p.CountRank(1))
}

func TestGoFishPlayer_RemoveAllOfRank(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 3, false))
	p.AddCard(NewCard(CardDesignClover, 5, false))

	removed := p.RemoveAllOfRank(3)
	assert.Len(t, removed, 2)
	assert.Equal(t, 1, p.GetCardsSize())
	assert.Equal(t, 5, p.GetCard(0).GetValue())
}

func TestGoFishPlayer_RemoveAllOfRank_NoneFound(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 3, false))

	removed := p.RemoveAllOfRank(5)
	assert.Len(t, removed, 0)
	assert.Equal(t, 1, p.GetCardsSize())
}

func TestGoFishPlayer_GetDistinctRanks(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 3, false))
	p.AddCard(NewCard(CardDesignClover, 7, false))

	ranks := p.GetDistinctRanks()
	assert.Len(t, ranks, 2)
	assert.Contains(t, ranks, 3)
	assert.Contains(t, ranks, 7)
}

func TestGoFishPlayer_JSON(t *testing.T) {
	p := NewGoFishPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddBook([]*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 3, false),
	})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored GoFishPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Equal(t, 1, restored.GetBookCount())
	assert.Equal(t, 5, restored.GetCard(0).GetValue())
}

func TestGoFishPlayer_JSON_Empty(t *testing.T) {
	data := []byte(`{"gp":null}`)
	var p GoFishPlayer
	require.NoError(t, json.Unmarshal(data, &p))
	assert.Equal(t, 0, p.GetBookCount())
	assert.Equal(t, 0, p.GetCardsSize())
}
