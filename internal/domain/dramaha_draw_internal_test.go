package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------------
// Draw round -- the phase Dramaha adds between the flop betting and the turn.
// The clone (Omaha) has no draw round at all, so none of these were inherited.
// --------------------------------------------------------------------------

// dramahaCardID identifies a card within one deck; suit+rank is unique, so it
// survives the copies that ReplaceCard/HoleCardsCopy make.
func dramahaCardID(c *Card) [2]int {
	if c == nil {
		return [2]int{-1, -1}
	}
	return [2]int{c.GetDesign(), c.GetValue()}
}

func dramahaHandIDs(p *DramahaPlayer) [][2]int {
	ids := make([][2]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		ids = append(ids, dramahaCardID(p.GetCard(i)))
	}
	return ids
}

// dealDramahaHole replaces a seat's hand with the given cards.
func dealDramahaHole(p *DramahaPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// diamondHole is a plain diamond flush -- deliberately not consecutive, so the
// draw side reads Flush and not Straight Flush. It holds only diamonds, so an
// unshuffled deck (which deals spades first) can never hand back a card the
// seat already holds.
func diamondHole() []*Card {
	return []*Card{
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignDiamond, 13, false),
	}
}

// newDramahaAtDraw parks a table in the draw round with seat 0 (the human)
// still to act and every other seat already done. The deck is left unshuffled
// so the replacement cards are known: spade A, spade 2, spade 3, ...
func newDramahaAtDraw() *Dramaha {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
		dealDramahaHole(p, diamondHole()...)
	}
	o.phase = DramahaPhaseDraw
	o.drawnFlags = []bool{false, true, true, true}
	o.communityCards = dramahaTestFlop()
	return o
}

// dramahaTestFlop is a rainbow flop that pairs nothing and completes nothing.
func dramahaTestFlop() []*Card {
	return []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 8, false),
	}
}

func TestDramahaDraw_ReplacesExactlyTheNamedCards(t *testing.T) {
	o := newDramahaAtDraw()
	before := dramahaHandIDs(o.players[0])

	require.NoError(t, o.Draw(0, []int{0, 2}))

	after := dramahaHandIDs(o.players[0])
	require.Len(t, after, DramahaHoleCards, "a draw must never change the hand size")

	// The two named positions took the top two cards off the deck, in order.
	assert.Equal(t, [2]int{CardDesignSpade, 1}, after[0], "position 0 must take the first card off the deck")
	assert.Equal(t, [2]int{CardDesignSpade, 2}, after[2], "position 2 must take the second card off the deck")

	// Everything else is untouched -- not re-drawn, not reordered.
	assert.Equal(t, before[1], after[1])
	assert.Equal(t, before[3], after[3])
	assert.Equal(t, before[4], after[4])
	assert.NotEqual(t, before[0], after[0])
	assert.NotEqual(t, before[2], after[2])
}

func TestDramahaDraw_EmptyIndicesStandsPat(t *testing.T) {
	for _, tc := range []struct {
		name    string
		indices []int
	}{
		{"nil", nil},
		{"empty", []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := newDramahaAtDraw()
			before := dramahaHandIDs(o.players[0])

			require.NoError(t, o.Draw(0, tc.indices))

			assert.Equal(t, before, dramahaHandIDs(o.players[0]), "standing pat must not touch the hand")
			assert.True(t, o.drawnFlags[0], "standing pat still uses up the seat's one draw")
		})
	}
}

func TestDramahaDraw_RejectsDuplicateIndex(t *testing.T) {
	o := newDramahaAtDraw()
	before := dramahaHandIDs(o.players[0])

	err := o.Draw(0, []int{1, 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, before, dramahaHandIDs(o.players[0]), "a rejected draw must leave the hand alone")
	assert.False(t, o.drawnFlags[0], "a rejected draw must not use up the seat's draw")
}

func TestDramahaDraw_RejectsIndexOutOfRange(t *testing.T) {
	for _, tc := range []struct {
		name    string
		indices []int
	}{
		{"below zero", []int{-1}},
		{"past the last card", []int{DramahaHoleCards}},
		{"valid then invalid", []int{0, 99}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := newDramahaAtDraw()
			before := dramahaHandIDs(o.players[0])

			err := o.Draw(0, tc.indices)

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidPlay)
			assert.Equal(t, before, dramahaHandIDs(o.players[0]),
				"the whole draw is rejected before any card is replaced")
			assert.False(t, o.drawnFlags[0])
		})
	}
}

func TestDramahaDraw_RejectsMoreCardsThanTheHandHolds(t *testing.T) {
	o := newDramahaAtDraw()

	err := o.Draw(0, []int{0, 1, 2, 3, 4, 0})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.False(t, o.drawnFlags[0])
}

func TestDramahaDraw_RejectsASecondDraw(t *testing.T) {
	o := newDramahaAtDraw()
	// Leave seat 1 pending so the round -- and with it the phase -- stays open;
	// otherwise the second call would be refused for the wrong reason.
	o.drawnFlags[1] = false
	require.NoError(t, o.Draw(0, []int{0}))
	require.Equal(t, DramahaPhaseDraw, o.phase)
	after := dramahaHandIDs(o.players[0])

	err := o.Draw(0, []int{1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Equal(t, after, dramahaHandIDs(o.players[0]),
		"the second draw must not reach the hand -- one exchange per round")
}

func TestDramahaDraw_RejectsUnknownSeat(t *testing.T) {
	for _, idx := range []int{-1, 4, 99} {
		o := newDramahaAtDraw()
		err := o.Draw(idx, []int{0})
		require.Error(t, err, "seat %d", idx)
		assert.ErrorIs(t, err, ErrInvalidPlay)
	}
}

func TestDramahaDraw_RejectedOutsideTheDrawPhase(t *testing.T) {
	for _, phase := range []int{
		DramahaPhaseInit, DramahaPhasePreFlop, DramahaPhaseFlop,
		DramahaPhaseTurn, DramahaPhaseRiver, DramahaPhaseShowdown, DramahaPhaseEnd,
	} {
		o := newDramahaAtDraw()
		o.phase = phase
		before := dramahaHandIDs(o.players[0])

		err := o.Draw(0, []int{0})

		require.Error(t, err, "phase %d", phase)
		assert.ErrorIs(t, err, ErrWrongPhase, "phase %d", phase)
		assert.Equal(t, before, dramahaHandIDs(o.players[0]), "phase %d", phase)
	}
}

func TestDramahaDraw_BettingIsRejectedDuringTheDrawRound(t *testing.T) {
	o := newDramahaAtDraw()
	o.currentTurn = 0
	o.actedFlags = []bool{false, true, true, true}

	err := o.PlayerAction(DramahaActionBet, 20, 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase,
		"the draw round has no betting; the turn round that follows does")
}

func TestDramahaDraw_FlopBettingOpensTheDrawRound(t *testing.T) {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
		dealDramahaHole(p, diamondHole()...)
	}
	o.phase = DramahaPhaseFlop
	o.communityCards = dramahaTestFlop()
	o.actedFlags = make([]bool, 4)

	o.advancePhase()

	assert.Equal(t, DramahaPhaseDraw, o.phase, "the flop is followed by the draw round, not the turn")
	assert.Len(t, o.communityCards, 3, "no turn card comes out until the draw round is over")
	require.Len(t, o.drawnFlags, o.GetPlayerCnt())
	assert.False(t, o.drawnFlags[0], "the human still has to draw")
	for i := 1; i < o.GetPlayerCnt(); i++ {
		assert.True(t, o.drawnFlags[i], "CPU seat %d draws automatically", i)
	}
}

func TestDramahaDraw_FoldedSeatIsPreMarkedDone(t *testing.T) {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
		dealDramahaHole(p, diamondHole()...)
	}
	o.players[2].SetFolded(true)
	foldedHand := dramahaHandIDs(o.players[2])
	o.phase = DramahaPhaseFlop
	o.communityCards = dramahaTestFlop()
	o.actedFlags = make([]bool, 4)

	o.advancePhase()

	assert.True(t, o.drawnFlags[2], "a folded seat is done before the round starts")
	assert.Equal(t, foldedHand, dramahaHandIDs(o.players[2]),
		"a folded seat must not be dealt replacement cards")

	err := o.Draw(2, []int{0})
	require.Error(t, err, "a folded seat cannot draw")
}

func TestDramahaDraw_PhaseAdvancesOnlyWhenEverySeatHasDrawn(t *testing.T) {
	// Two seats still to draw: the phase must hold until both are done.
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
		dealDramahaHole(p, diamondHole()...)
	}
	o.phase = DramahaPhaseDraw
	o.drawnFlags = []bool{false, false, true, true}
	o.communityCards = dramahaTestFlop()

	require.NoError(t, o.Draw(0, []int{0}))
	assert.Equal(t, DramahaPhaseDraw, o.phase, "one seat left: the round is not over")
	assert.Len(t, o.communityCards, 3, "the turn card must not come out early")

	require.NoError(t, o.Draw(1, []int{0}))
	assert.Equal(t, DramahaPhaseTurn, o.phase, "the last draw closes the round")
	assert.Len(t, o.communityCards, 4, "the turn card comes out with the phase")
}

func TestDramahaDraw_DrawRoundIsFollowedByTheTurn(t *testing.T) {
	o := newDramahaAtDraw()

	o.advancePhase()

	assert.Equal(t, DramahaPhaseTurn, o.phase)
	assert.Len(t, o.communityCards, 4)
}

func TestDramahaDraw_ExchangeChangesTheDrawHand(t *testing.T) {
	o := newDramahaAtDraw()
	// A diamond flush before the draw...
	require.Equal(t, PokerHandFlush, o.players[0].EvalDrawHand())

	// ...broken by swapping one diamond for the spade on top of the deck.
	require.NoError(t, o.Draw(0, []int{0}))

	assert.Equal(t, PokerHandHighCard, o.players[0].EvalDrawHand(),
		"the draw hand is read off the hole cards, so an exchange must change it")
}

func TestDramahaAutoDrawForCPUs_LeavesTheHumanAlone(t *testing.T) {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
		dealDramahaHole(p, diamondHole()...)
	}
	o.phase = DramahaPhaseDraw
	o.drawnFlags = make([]bool, 4)
	humanHand := dramahaHandIDs(o.players[0])

	o.autoDrawForCPUs()

	assert.False(t, o.drawnFlags[0], "the human draws for themselves")
	assert.Equal(t, humanHand, dramahaHandIDs(o.players[0]))
	for i := 1; i < 4; i++ {
		assert.True(t, o.drawnFlags[i], "CPU seat %d should have drawn", i)
	}
}

func TestDramahaCPUDiscards_KeepsThePairedCards(t *testing.T) {
	t.Run("keeps a pair and throws the rest", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignSpade, 9, false),
		}
		assert.Equal(t, []int{2, 3, 4}, dramahaCPUDiscards(cards))
	})

	t.Run("with nothing paired it keeps the two highest", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 1, false), // ace is the high card, not the low one
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 13, false),
			NewCard(CardDesignSpade, 4, false),
		}
		assert.Equal(t, []int{0, 2, 4}, dramahaCPUDiscards(cards))
	})

	t.Run("a hand that is not five cards is left alone", func(t *testing.T) {
		assert.Nil(t, dramahaCPUDiscards([]*Card{NewCard(CardDesignSpade, 2, false)}))
	})
}
