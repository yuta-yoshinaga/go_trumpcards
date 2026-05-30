package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestIrishPoker() *Pineapple {
	cfg := DefaultPineappleConfig()
	players := NewPineapplePlayersForTable(HoldemTableSize4)
	return NewIrishPoker(NewTrumpCards(0), players, cfg)
}

func TestNewIrishPoker(t *testing.T) {
	p := newTestIrishPoker()
	assert.NotNil(t, p)
	assert.True(t, p.IsDiscardAfterFlopBetting())
	assert.Equal(t, 4, p.GetInitialDealCount())
}

func TestNewDefaultIrishPoker(t *testing.T) {
	p := NewDefaultIrishPoker()
	assert.NotNil(t, p)
	assert.True(t, p.IsDiscardAfterFlopBetting())
	assert.Equal(t, 4, p.GetInitialDealCount())
}

func TestIrishPoker_Reset_Deals4Cards(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())

	for _, pl := range p.players {
		assert.Equal(t, 4, pl.GetCardsSize(), "Irish Poker should deal 4 hole cards")
	}
}

func TestIrishPoker_DiscardTwoCards(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())

	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}
	p.phase = PineapplePhaseDiscard
	p.discardDone = make([]bool, len(p.players))
	for i := 1; i < len(p.players); i++ {
		p.discardDone[i] = true
	}
	p.players[0].Reset()
	p.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	p.players[0].AddCard(NewCard(CardDesignHeart, 6, false))
	p.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	p.players[0].AddCard(NewCard(CardDesignHeart, 8, false))

	require.NoError(t, p.DiscardCard(0))
	assert.Equal(t, 3, p.players[0].GetCardsSize(), "should have 3 cards after first discard")
	assert.False(t, p.discardDone[0], "not done yet after first discard")
	assert.Equal(t, PineapplePhaseDiscard, p.GetPhase(), "should still be in discard phase")

	require.NoError(t, p.DiscardCard(0))
	assert.Equal(t, 2, p.players[0].GetCardsSize(), "should have 2 cards after second discard")
	assert.True(t, p.discardDone[0], "done after second discard")
	assert.Equal(t, PineapplePhaseTurn, p.GetPhase(), "should advance to turn after all discards done")
	assert.Equal(t, 4, len(p.GetCommunityCards()), "turn card must be dealt")
}

func TestIrishPoker_CPUAutoDiscard(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())

	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}

	for _, pl := range p.players {
		pl.Reset()
		pl.AddCard(NewCard(CardDesignHeart, 5, false))
		pl.AddCard(NewCard(CardDesignHeart, 6, false))
		pl.AddCard(NewCard(CardDesignHeart, 7, false))
		pl.AddCard(NewCard(CardDesignHeart, 8, false))
	}

	p.enterDiscardPhase()

	for i, pl := range p.players {
		if pl.GetIsHuman() {
			assert.Equal(t, 4, pl.GetCardsSize(), "human should still have 4 cards")
			assert.False(t, p.discardDone[i])
		} else {
			assert.Equal(t, 2, pl.GetCardsSize(), "CPU should have 2 cards after auto-discard")
			assert.True(t, p.discardDone[i])
		}
	}
}

func TestIrishPoker_AllInAutoDiscard(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())

	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}

	for i, pl := range p.players {
		pl.Reset()
		pl.AddCard(NewCard(CardDesignHeart, 5, false))
		pl.AddCard(NewCard(CardDesignHeart, 6, false))
		pl.AddCard(NewCard(CardDesignHeart, 7, false))
		pl.AddCard(NewCard(CardDesignHeart, 8, false))
		if i == 0 {
			pl.SetAllIn(true)
		}
	}

	p.enterDiscardPhase()

	assert.Equal(t, 2, p.players[0].GetCardsSize(), "all-in player should auto-discard to 2 cards")
	assert.True(t, p.discardDone[0])
}

func TestIrishPoker_JSON_RoundTrip(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Pineapple
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.IsDiscardAfterFlopBetting())
	assert.Equal(t, 4, restored.GetInitialDealCount())
}

func TestPineapple_Default_InitialDealCount(t *testing.T) {
	p := newTestPineapple()
	assert.Equal(t, 3, p.GetInitialDealCount())
}

func TestPineapple_JSON_OldData_DefaultsTo3(t *testing.T) {
	raw := `{"tc":{},"pl":[],"cc":[],"pt":0,"sp":[],"di":0,"ct":0,"ph":0,"cf":{},"ge":false,"lb":0,"mr":0,"rc":0,"af":[],"rr":[],"ca":[],"sc":[],"al":[],"dd":[]}`
	var p Pineapple
	require.NoError(t, json.Unmarshal([]byte(raw), &p))
	assert.Equal(t, 3, p.GetInitialDealCount(), "missing idc field should default to 3")
}

func TestIrishPoker_AdvancePhase_PreFlopToFlop_NoDiscard(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())
	p.phase = PineapplePhasePreFlop

	p.advancePhase()

	assert.Equal(t, PineapplePhaseFlop, p.GetPhase(), "should advance to flop betting (not discard) in Irish mode")
	assert.Equal(t, 3, len(p.GetCommunityCards()), "flop must be dealt")
}

func TestIrishPoker_AdvancePhase_FlopToDiscard(t *testing.T) {
	p := newTestIrishPoker()
	require.NoError(t, p.Reset())
	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}
	p.phase = PineapplePhaseFlop
	for _, pl := range p.players {
		pl.Reset()
		pl.AddCard(NewCard(CardDesignHeart, 5, false))
		pl.AddCard(NewCard(CardDesignHeart, 6, false))
		pl.AddCard(NewCard(CardDesignHeart, 7, false))
		pl.AddCard(NewCard(CardDesignHeart, 8, false))
	}

	p.advancePhase()

	assert.Equal(t, PineapplePhaseDiscard, p.GetPhase())
	assert.True(t, p.IsDiscardPhase())
}
