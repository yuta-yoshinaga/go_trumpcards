//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// getBadugiPlayerCards reads all cards from a BadugiPlayer via GetCardsSize/GetCard
// since Player does not expose a slice accessor.
func getBadugiPlayerCards(p *domain.BadugiPlayer) []*domain.Card {
	n := p.GetCardsSize()
	out := make([]*domain.Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

func newBadugiTestPlayer(cards ...[2]int) *domain.BadugiPlayer {
	p := domain.NewBadugiPlayer(true, domain.BadugiStyleBalanced)
	for _, pair := range cards {
		p.AddCard(domain.NewCard(pair[0], pair[1], true))
	}
	return p
}

func TestBadugiPlayer_EvalHand(t *testing.T) {
	S, C, H, D := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond

	t.Run("perfect Badugi returns Size 4", func(t *testing.T) {
		p := newBadugiTestPlayer([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4})
		assert.Equal(t, 4, p.EvalHand())
		assert.Equal(t, 4, p.GetHandRank())
		assert.Equal(t, "Badugi", p.GetHandName())
	})

	t.Run("three-card hand returns Size 3", func(t *testing.T) {
		p := newBadugiTestPlayer([2]int{S, 1}, [2]int{H, 1}, [2]int{D, 3}, [2]int{C, 4})
		assert.Equal(t, 3, p.EvalHand())
		assert.Equal(t, "3-card", p.GetHandName())
	})

	t.Run("unknown hand name before eval", func(t *testing.T) {
		p := domain.NewBadugiPlayer(false, domain.BadugiStyleConservative)
		assert.Equal(t, "Unknown", p.GetHandName())
	})
}

func TestBadugiPlayer_ExchangeCard(t *testing.T) {
	S, C, H, D := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond
	p := newBadugiTestPlayer([2]int{S, 1}, [2]int{S, 2}, [2]int{S, 3}, [2]int{S, 4})
	// Replace position 0 with a heart 5 → now S has three (2,3,4) + H5 → best size 2
	p.ExchangeCard(0, domain.NewCard(H, 5, true))
	_ = p.EvalHand()
	assert.Equal(t, 2, p.GetHandRank())

	// Out-of-range exchange is a silent no-op and does not panic.
	p.ExchangeCard(99, domain.NewCard(D, 7, true))
	p.ExchangeCard(-1, domain.NewCard(C, 8, true))
	assert.Equal(t, 4, len(getBadugiPlayerCards(p)))
}

func TestBadugiPlayer_DrawCounters(t *testing.T) {
	p := domain.NewBadugiPlayer(false, domain.BadugiStyleAggressive)
	p.SetDrawCount(2)
	p.AddToTotalDrawCount(2)
	p.AddToTotalDrawCount(1)
	assert.Equal(t, 2, p.GetDrawCount())
	assert.Equal(t, 3, p.GetTotalDrawCount())

	p.ResetDrawCounters()
	assert.Equal(t, 0, p.GetDrawCount())
	assert.Equal(t, 0, p.GetTotalDrawCount())
}

func TestBadugiPlayer_Metadata(t *testing.T) {
	p := domain.NewBadugiPlayer(false, domain.BadugiStyleBluffer)
	assert.False(t, p.GetIsHuman())
	assert.Equal(t, domain.BadugiStyleBluffer, p.GetPlayStyle())
	assert.Equal(t, "Bluffer", p.GetPlayStyleName())

	human := domain.NewBadugiPlayer(true, domain.BadugiStyleBalanced)
	assert.True(t, human.GetIsHuman())
}

func TestBadugiPlayer_GetComparisonCards(t *testing.T) {
	S, C, H, D := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond
	p := newBadugiTestPlayer([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4})
	_ = p.EvalHand()
	cards := p.GetComparisonCards()
	assert.Len(t, cards, 4)
	// Cached Cards are sorted descending — index 0 is the top (4), index 3 is the Ace.
	assert.Equal(t, 4, cards[0].GetValue())
	assert.Equal(t, 1, cards[3].GetValue())
}

func TestBadugiPlayer_RoundTripJSON(t *testing.T) {
	S, H, D, C := domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignClover
	orig := newBadugiTestPlayer([2]int{S, 1}, [2]int{H, 2}, [2]int{D, 3}, [2]int{C, 4})
	orig.AddChips(500)
	orig.SetDrawCount(2)
	orig.AddToTotalDrawCount(2)

	data, err := json.Marshal(orig)
	assert.NoError(t, err)

	round := domain.NewBadugiPlayer(false, domain.BadugiStyleConservative)
	assert.NoError(t, json.Unmarshal(data, round))
	assert.Equal(t, orig.GetChips(), round.GetChips())
	assert.Equal(t, orig.GetDrawCount(), round.GetDrawCount())
	assert.Equal(t, orig.GetTotalDrawCount(), round.GetTotalDrawCount())
	assert.Equal(t, orig.GetIsHuman(), round.GetIsHuman())
	assert.Equal(t, orig.GetPlayStyle(), round.GetPlayStyle())
	// Cards survive the round-trip.
	assert.Equal(t, orig.GetCardsSize(), round.GetCardsSize())
}
