//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewTongitsPlayer(t *testing.T) {
	p := domain.NewTongitsPlayer(true)

	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCumulativeScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Empty(t, p.GetMelds())
}

func TestTongitsPlayer_Cards(t *testing.T) {
	p := domain.NewTongitsPlayer(false)
	ace := domain.NewCard(domain.CardDesignSpade, 1, false)
	nine := domain.NewCard(domain.CardDesignHeart, 9, false)
	king := domain.NewCard(domain.CardDesignClover, 13, false)
	p.AddCard(ace)
	p.AddCard(nine)
	p.AddCard(king)

	assert.Equal(t, 3, p.GetCardsSize())
	assert.Same(t, ace, p.GetCard(0))
	assert.Same(t, nine, p.RemoveCard(1))
	assert.Equal(t, 2, p.GetCardsSize())
	assert.Same(t, king, p.GetCard(1))
	assert.Nil(t, p.RemoveCard(10))
}

func TestTongitsPlayer_RemainingCardPoints(t *testing.T) {
	assert.Equal(t, 1, domain.TongitsCardValue(domain.NewCard(domain.CardDesignSpade, 1, false)))
	assert.Equal(t, 7, domain.TongitsCardValue(domain.NewCard(domain.CardDesignHeart, 7, false)))
	assert.Equal(t, 10, domain.TongitsCardValue(domain.NewCard(domain.CardDesignClover, 11, false)))
	assert.Equal(t, 10, domain.TongitsCardValue(domain.NewCard(domain.CardDesignDiamond, 13, false)))
	assert.Zero(t, domain.TongitsCardValue(nil))
}

func TestTongitsPlayer_PublicMelds(t *testing.T) {
	p := domain.NewTongitsPlayer(true)
	first := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignSpade, 4, false),
	}
	added := domain.NewCard(domain.CardDesignSpade, 5, false)
	p.AppendMeld(first)
	p.AddCardToMeld(0, added)

	assert.Len(t, p.GetMelds(), 1)
	assert.Len(t, p.GetMeld(0), 4)
	assert.Same(t, added, p.GetMeld(0)[3])
	assert.Nil(t, p.GetMeld(-1))
	assert.Nil(t, p.GetMeld(1))

	p.ClearMelds()
	assert.Empty(t, p.GetMelds())
}

func TestTongitsPlayer_ResetRound(t *testing.T) {
	p := domain.NewTongitsPlayer(true)
	p.SetRoundScore(30)
	p.SetCumulativeScore(60)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p.AppendMeld([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 60, p.GetCumulativeScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Empty(t, p.GetMelds())
	assert.False(t, p.GetIsFinished())
}

func TestTongitsPlayer_JSONRoundtrip(t *testing.T) {
	p := domain.NewTongitsPlayer(true)
	p.SetRoundScore(15)
	p.SetCumulativeScore(60)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p.AppendMeld([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored domain.TongitsPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 15, restored.GetRoundScore())
	assert.Equal(t, 60, restored.GetCumulativeScore())
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Len(t, restored.GetMeld(0), 1)
}

func TestTongitsPlayer_UnmarshalNilGamePlayer(t *testing.T) {
	var p domain.TongitsPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))

	assert.False(t, p.GetIsHuman())
	assert.Empty(t, p.GetMelds())
}
