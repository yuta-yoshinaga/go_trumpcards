package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPineapplePlayer(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, "TAG", p.GetPlayStyleName())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestPineapplePlayer_HUDStats(t *testing.T) {
	p := NewPineapplePlayer(false, HoldemStyleLAG)

	assert.Equal(t, 0, p.GetVPIP())
	assert.Equal(t, 0, p.GetPFR())
	assert.Equal(t, 0, p.GetThreeBet())
	assert.Equal(t, "-", p.GetAFDisplay())

	p.IncrementTotalHands()
	p.IncrementTotalHands()
	p.IncrementVPIP()
	p.IncrementPFR()
	assert.Equal(t, 50, p.GetVPIP())
	assert.Equal(t, 50, p.GetPFR())

	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBet()
	assert.Equal(t, 50, p.GetThreeBet())

	p.IncrementPostFlopBetRaise()
	p.IncrementPostFlopBetRaise()
	assert.Equal(t, "∞", p.GetAFDisplay())

	p.IncrementPostFlopCall()
	assert.Equal(t, "2.0", p.GetAFDisplay())
}

func TestPineapplePlayer_EvalBestHand(t *testing.T) {
	t.Run("2 hole cards + 5 community = standard holdem eval", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace spades
		p.AddCard(NewCard(CardDesignSpade, 13, false)) // King spades

		comm := []*Card{
			NewCard(CardDesignSpade, 12, false), // Q spades
			NewCard(CardDesignSpade, 11, false), // J spades
			NewCard(CardDesignSpade, 10, false), // 10 spades
			NewCard(CardDesignHeart, 2, false),  // 2 hearts
			NewCard(CardDesignHeart, 3, false),  // 3 hearts
		}
		rank := p.EvalBestHand(comm)
		assert.Equal(t, PokerHandRoyalFlush, rank)
		assert.NotNil(t, p.GetBestHand())
	})

	t.Run("3 hole cards + 3 community = best 5 of 6", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace spades
		p.AddCard(NewCard(CardDesignSpade, 13, false)) // King spades
		p.AddCard(NewCard(CardDesignHeart, 2, false))  // 2 hearts (weak)

		comm := []*Card{
			NewCard(CardDesignSpade, 12, false), // Q spades
			NewCard(CardDesignSpade, 11, false), // J spades
			NewCard(CardDesignSpade, 10, false), // 10 spades
		}
		rank := p.EvalBestHand(comm)
		assert.Equal(t, PokerHandRoyalFlush, rank)
	})

	t.Run("less than 5 total cards returns HighCard", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignSpade, 13, false))

		rank := p.EvalBestHand([]*Card{NewCard(CardDesignHeart, 2, false)})
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, p.GetBestHand())
	})
}

func TestPineapplePlayer_GetComparisonCards(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))

	comm := []*Card{
		NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 11, false), NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 3, false),
	}
	p.EvalBestHand(comm)

	cards := p.GetComparisonCards()
	assert.Equal(t, 5, len(cards))
	// Verify it's a copy
	cards[0] = nil
	assert.NotNil(t, p.GetComparisonCards()[0])
}

func TestPineapplePlayer_JSON(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleLAG)
	p.SetChips(500)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	p.IncrementTotalHands()
	p.IncrementVPIP()

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	p2 := &PineapplePlayer{}
	err = json.Unmarshal(data, p2)
	assert.NoError(t, err)

	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, HoldemStyleLAG, p2.GetPlayStyle())
	assert.Equal(t, 500, p2.GetChips())
	assert.Equal(t, 1, p2.GetTotalHands())
	assert.Equal(t, 1, p2.GetVPIPCount())
	assert.Equal(t, 2, p2.GetCardsSize())
}
