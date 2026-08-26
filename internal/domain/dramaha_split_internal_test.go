package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------------
// Showdown -- the pot ALWAYS splits 50:50 between the Omaha side (exactly two
// hole cards + exactly three board cards) and the draw side (the five hole
// cards as they are). The clone's "no qualifying low, high scoops" branch is
// gone: five cards always rank, so the draw side always has a winner.
// --------------------------------------------------------------------------

// dramahaTotalChips is the invariant every showdown must preserve: chips are
// only ever moved between the pot and the seats, never created or destroyed.
// Per-seat assertions cannot see money that appears or vanishes; this can.
func dramahaTotalChips(o *Dramaha) int {
	total := o.pot
	for _, p := range o.players {
		total += p.GetChips()
	}
	return total
}

// splitBoard pairs nothing and completes nothing on its own: K-Q-J of spades
// plus two rags.
func splitBoard() []*Card {
	return []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 7, false),
	}
}

// newDramahaAtShowdown seats four players, folds seats 2 and 3, and parks the
// table on the river board above with `pot` in the middle.
func newDramahaAtShowdown(pot int) *Dramaha {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
	}
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.phase = DramahaPhaseShowdown
	o.pot = pot
	o.communityCards = splitBoard()
	o.players[2].SetFolded(true)
	o.players[3].SetFolded(true)
	return o
}

// heartFlushHole wins the draw side (a flush of five hearts) and loses the
// Omaha side (it can pair nothing on the board and cannot reach five hearts
// with only one heart out there).
func heartFlushHole() []*Card {
	return []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false),
	}
}

// ragHole loses both sides against anything. It pairs nothing on the board,
// reaches no straight with it (the board's only run is K-Q-J, which would need
// 9 and 10 in the hole), and cannot reach a flush (never two cards of a suit
// the board holds three of), so it is high card on both sides. Its cards are
// disjoint from twoKingsHole so the two can sit at the same table.
func ragHole() []*Card {
	return []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignDiamond, 9, false),
	}
}

// twoKingsHole wins the Omaha side (two kings + the board's king-queen-jack is
// trip kings) and loses the draw side (as dealt it is only a pair of kings).
func twoKingsHole() []*Card {
	return []*Card{
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignHeart, 8, false),
	}
}

// TestDramahaShowdown_FixtureSplitsTheTwoSides is the negative control for the
// split tests below: it proves the fixture really does put the two winners on
// different sides, so a passing split test is not passing by accident.
func TestDramahaShowdown_FixtureSplitsTheTwoSides(t *testing.T) {
	o := newDramahaAtShowdown(200)
	dealDramahaHole(o.players[0], heartFlushHole()...)
	dealDramahaHole(o.players[1], twoKingsHole()...)

	assert.Equal(t, PokerHandHighCard, o.players[0].EvalBestHand(o.communityCards))
	assert.Equal(t, PokerHandThreeOfAKind, o.players[1].EvalBestHand(o.communityCards))
	assert.Equal(t, PokerHandFlush, o.players[0].EvalDrawHand())
	assert.Equal(t, PokerHandOnePair, o.players[1].EvalDrawHand())
}

func TestDramahaShowdown_PotAlwaysSplitsBetweenTheTwoSides(t *testing.T) {
	o := newDramahaAtShowdown(200)
	dealDramahaHole(o.players[0], heartFlushHole()...)
	dealDramahaHole(o.players[1], twoKingsHole()...)
	before := dramahaTotalChips(o)

	o.resolveShowdown()

	assert.Equal(t, before, dramahaTotalChips(o), "chips must be conserved across the showdown")
	assert.Equal(t, 0, o.pot, "a resolved showdown leaves nothing in the middle")
	assert.Equal(t, 1100, o.players[0].GetChips(), "the draw winner takes half")
	assert.Equal(t, 1100, o.players[1].GetChips(), "the Omaha winner takes the other half")

	byIdx := map[int]HoldemResult{}
	for _, r := range o.roundResults {
		byIdx[r.PlayerIdx] = r
	}
	require.Len(t, byIdx, 2, "only the seats that saw the showdown get a result")

	assert.Equal(t, 0, byIdx[0].HiWonAmount, "the flush loses the Omaha half")
	assert.Equal(t, 100, byIdx[0].LowWonAmount, "the flush wins the draw half")
	assert.Equal(t, 100, byIdx[1].HiWonAmount, "trip kings win the Omaha half")
	assert.Equal(t, 0, byIdx[1].LowWonAmount, "a pair loses the draw half")

	for idx, r := range byIdx {
		assert.True(t, r.LowQualifies, "seat %d: five cards always rank, so the draw side always counts", idx)
		assert.Len(t, r.LowBestHand, DramahaHoleCards, "seat %d: the draw hand is the whole hole", idx)
	}
}

func TestDramahaShowdown_SingleWinnerTakesBothHalves(t *testing.T) {
	o := newDramahaAtShowdown(200)
	// Two kings AND a heart flush would be impossible; instead give the human
	// the trip kings and the opponent a hand that loses both ways.
	dealDramahaHole(o.players[0], twoKingsHole()...)
	dealDramahaHole(o.players[1], ragHole()...)
	require.Equal(t, PokerHandHighCard, o.players[1].EvalBestHand(o.communityCards))
	require.Equal(t, PokerHandHighCard, o.players[1].EvalDrawHand())
	before := dramahaTotalChips(o)

	o.resolveShowdown()

	assert.Equal(t, before, dramahaTotalChips(o), "chips must be conserved across a scoop")
	assert.Equal(t, 0, o.pot)
	assert.Equal(t, 1200, o.players[0].GetChips(), "one player winning both sides takes the whole pot")
	assert.Equal(t, 1000, o.players[1].GetChips())

	for _, r := range o.roundResults {
		if r.PlayerIdx == 0 {
			assert.Equal(t, 100, r.HiWonAmount)
			assert.Equal(t, 100, r.LowWonAmount)
			assert.Equal(t, 200, r.WonAmount)
		}
	}
}

func TestDramahaShowdown_OddChipGoesToTheOmahaSide(t *testing.T) {
	o := newDramahaAtShowdown(101)
	dealDramahaHole(o.players[0], heartFlushHole()...)
	dealDramahaHole(o.players[1], twoKingsHole()...)
	before := dramahaTotalChips(o)

	o.resolveShowdown()

	assert.Equal(t, before, dramahaTotalChips(o), "the odd chip must not be dropped on the table")
	assert.Equal(t, 1050, o.players[0].GetChips(), "the draw side rounds down")
	assert.Equal(t, 1051, o.players[1].GetChips(), "the odd chip goes to the Omaha side")
}

func TestDramahaShowdown_ChipsAreConservedWhenTheHumanLoses(t *testing.T) {
	// A losing human stops short of finalizeShowdown so the muck prompt can be
	// offered; the chips only balance again once the hand is closed out.
	o := newDramahaAtShowdown(200)
	dealDramahaHole(o.players[0], ragHole()...)
	dealDramahaHole(o.players[1], twoKingsHole()...)
	require.Equal(t, PokerHandHighCard, o.players[0].EvalBestHand(o.communityCards))
	require.Equal(t, PokerHandHighCard, o.players[0].EvalDrawHand())
	before := dramahaTotalChips(o)

	o.resolveShowdown()
	assert.Equal(t, DramahaPhaseShowdown, o.phase, "a losing human is offered the muck prompt")
	assert.True(t, o.IsMuckAvailable())

	require.NoError(t, o.ShowHand())

	assert.Equal(t, before, dramahaTotalChips(o), "chips must be conserved once the hand is closed")
	assert.Equal(t, 0, o.pot)
	assert.Equal(t, 1200, o.players[1].GetChips())
	assert.Equal(t, DramahaPhaseEnd, o.phase)
}

func TestDramahaShowdown_EachSideCanTieAndSplitItsHalf(t *testing.T) {
	// Two flushes of the same five ranks in different suits: the draw side ties
	// on a flush and the Omaha side ties on the same king-high board play.
	o := newDramahaAtShowdown(200)
	dealDramahaHole(o.players[0],
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignClover, 9, false),
	)
	dealDramahaHole(o.players[1], heartFlushHole()...)
	require.Equal(t, PokerHandFlush, o.players[0].EvalDrawHand())
	require.Equal(t, PokerHandFlush, o.players[1].EvalDrawHand())
	require.Equal(t, PokerHandHighCard, o.players[0].EvalBestHand(o.communityCards))
	require.Equal(t, PokerHandHighCard, o.players[1].EvalBestHand(o.communityCards))
	before := dramahaTotalChips(o)

	o.resolveShowdown()

	assert.Equal(t, before, dramahaTotalChips(o))
	assert.Equal(t, 1100, o.players[0].GetChips(), "a tie on both sides pays each seat half the pot")
	assert.Equal(t, 1100, o.players[1].GetChips())
}

// --------------------------------------------------------------------------
// findDramahaDrawWinners
// --------------------------------------------------------------------------

func TestDramahaFindDrawWinners_ComparesRankFirst(t *testing.T) {
	o := newDramahaAtShowdown(100)
	dealDramahaHole(o.players[0], heartFlushHole()...) // flush
	dealDramahaHole(o.players[1], twoKingsHole()...)   // pair
	o.players[0].EvalDrawHand()
	o.players[1].EvalDrawHand()

	assert.Equal(t, []int{0}, o.findDramahaDrawWinners([]int{0, 1}))
}

func TestDramahaFindDrawWinners_BreaksARankTieOnTheHighCards(t *testing.T) {
	// Both hands are one pair. Rank alone would call this a tie and split the
	// draw half between a pair of aces and a pair of twos.
	o := newDramahaAtShowdown(100)
	dealDramahaHole(o.players[0],
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignSpade, 9, false),
	)
	dealDramahaHole(o.players[1],
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignHeart, 9, false),
	)
	require.Equal(t, PokerHandOnePair, o.players[0].EvalDrawHand())
	require.Equal(t, PokerHandOnePair, o.players[1].EvalDrawHand())

	assert.Equal(t, []int{0}, o.findDramahaDrawWinners([]int{0, 1}),
		"aces must beat twos at the same rank")
	assert.Equal(t, []int{0}, o.findDramahaDrawWinners([]int{1, 0}),
		"and must do so regardless of seat order")
}

func TestDramahaFindDrawWinners_TrueTieSplits(t *testing.T) {
	o := newDramahaAtShowdown(100)
	for _, idx := range []int{0, 1} {
		dealDramahaHole(o.players[idx],
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		)
		o.players[idx].EvalDrawHand()
	}

	assert.Equal(t, []int{0, 1}, o.findDramahaDrawWinners([]int{0, 1}))
}

func TestDramahaFindDrawWinners_SkipsFoldedSeats(t *testing.T) {
	o := newDramahaAtShowdown(100)
	dealDramahaHole(o.players[0], twoKingsHole()...)
	dealDramahaHole(o.players[2], heartFlushHole()...) // folded, and stronger
	o.players[0].EvalDrawHand()
	o.players[2].EvalDrawHand()

	assert.Equal(t, []int{0}, o.findDramahaDrawWinners([]int{0, 2}),
		"a folded seat must not win the draw half with a hand it threw away")
}
