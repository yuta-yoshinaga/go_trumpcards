package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestCrazyPineapple returns a Crazy Pineapple game ready for tests.
func newTestCrazyPineapple() *Pineapple {
	cfg := DefaultPineappleConfig()
	players := NewPineapplePlayersForTable(HoldemTableSize4)
	return NewCrazyPineapple(NewTrumpCards(0), players, cfg)
}

func TestNewCrazyPineapple(t *testing.T) {
	p := newTestCrazyPineapple()
	assert.NotNil(t, p)
	assert.True(t, p.IsDiscardAfterFlopBetting())
}

func TestNewDefaultCrazyPineapple(t *testing.T) {
	p := NewDefaultCrazyPineapple()
	assert.NotNil(t, p)
	assert.True(t, p.IsDiscardAfterFlopBetting())
}

func TestPineapple_Default_IsNotCrazy(t *testing.T) {
	p := newTestPineapple()
	assert.False(t, p.IsDiscardAfterFlopBetting())
}

// TestCrazyPineapple_AdvancePhase_PreFlopToFlop_NoDiscard asserts that in
// Crazy mode, ending the pre-flop betting round deals the flop and advances
// to the flop betting phase WITHOUT entering the discard phase.
func TestCrazyPineapple_AdvancePhase_PreFlopToFlop_NoDiscard(t *testing.T) {
	p := newTestCrazyPineapple()
	require.NoError(t, p.Reset())
	p.phase = PineapplePhasePreFlop

	p.advancePhase()

	assert.Equal(t, PineapplePhaseFlop, p.GetPhase(), "should advance to flop betting (not discard) in crazy mode")
	assert.Equal(t, 3, len(p.GetCommunityCards()), "flop must be dealt")
}

// TestCrazyPineapple_AdvancePhase_FlopToDiscard asserts that in Crazy mode,
// ending the flop betting round enters the discard phase.
func TestCrazyPineapple_AdvancePhase_FlopToDiscard(t *testing.T) {
	p := newTestCrazyPineapple()
	require.NoError(t, p.Reset())
	// Simulate: flop already dealt, now at flop betting
	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}
	p.phase = PineapplePhaseFlop
	// Reset players to 3 hole cards (no discard yet in crazy mode pre-discard)
	for _, pl := range p.players {
		pl.Reset()
		pl.AddCard(NewCard(CardDesignHeart, 5, false))
		pl.AddCard(NewCard(CardDesignHeart, 6, false))
		pl.AddCard(NewCard(CardDesignHeart, 7, false))
	}

	p.advancePhase()

	assert.Equal(t, PineapplePhaseDiscard, p.GetPhase(), "flop betting must transition to discard phase in crazy mode")
	assert.True(t, p.IsDiscardPhase())
}

// TestCrazyPineapple_AfterDiscard_DealsTurn asserts that completing the
// discard phase in Crazy mode deals the turn card and advances to the turn
// betting phase (not back to flop betting).
func TestCrazyPineapple_AfterDiscard_DealsTurn(t *testing.T) {
	p := newTestCrazyPineapple()
	require.NoError(t, p.Reset())
	// Flop cards already on the board, flop betting just ended.
	p.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
	}
	p.phase = PineapplePhaseDiscard
	p.discardDone = make([]bool, len(p.players))
	// Mark all CPUs done; human still needs to discard.
	for i := 1; i < len(p.players); i++ {
		p.discardDone[i] = true
	}
	p.players[0].Reset()
	p.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	p.players[0].AddCard(NewCard(CardDesignHeart, 6, false))
	p.players[0].AddCard(NewCard(CardDesignHeart, 7, false))

	require.NoError(t, p.DiscardCard(0))

	assert.Equal(t, PineapplePhaseTurn, p.GetPhase(), "crazy mode should advance to turn (not flop) after discard")
	assert.Equal(t, 4, len(p.GetCommunityCards()), "turn card must be dealt after crazy discard")
	assert.Equal(t, 2, p.players[0].GetCardsSize(), "human should retain two hole cards after discard")
}

// TestPineapple_Default_AfterDiscard_StaysFlopBetting is a regression guard:
// existing Pineapple behaviour (discard between flop deal and flop betting)
// must NOT be affected by the Crazy Pineapple changes.
func TestPineapple_Default_AfterDiscard_StaysFlopBetting(t *testing.T) {
	p := newTestPineapple() // default Pineapple (discardAfterFlopBetting == false)
	require.NoError(t, p.Reset())
	// Flop already dealt, now in discard phase (existing Pineapple flow).
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

	require.NoError(t, p.DiscardCard(0))

	assert.Equal(t, PineapplePhaseFlop, p.GetPhase(), "default pineapple must stay at flop betting after discard")
	assert.Equal(t, 3, len(p.GetCommunityCards()), "turn card must NOT be dealt in default pineapple after discard")
}

// TestCrazyPineapple_JSON_RoundTrip asserts the discardAfterFlopBetting flag
// survives snapshot + restore.
func TestCrazyPineapple_JSON_RoundTrip(t *testing.T) {
	p := newTestCrazyPineapple()
	require.NoError(t, p.Reset())
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Pineapple
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.IsDiscardAfterFlopBetting(), "crazy flag must persist across JSON round-trip")
}

func TestPineapple_Default_JSON_CrazyFlagFalse(t *testing.T) {
	p := newTestPineapple()
	require.NoError(t, p.Reset())
	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored Pineapple
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.False(t, restored.IsDiscardAfterFlopBetting())
}
