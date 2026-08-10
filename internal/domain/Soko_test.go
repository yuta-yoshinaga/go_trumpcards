//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fourFlushHand is a pair of kings that is ALSO four spades. Standard poker sees
// One Pair; Soko must see a four-card flush. Using one hand for both assertions
// is what makes the mode observable — a hand that only Soko can rank would pass
// even if the mode never reached the evaluator.
func fourFlushHand() (hole []*Card, door []*Card) {
	return []*Card{sp(2)}, []*Card{sp(5), sp(9), sp(13), he(13)}
}

func TestNewDefaultSoko_SetsModeOnGameAndPlayers(t *testing.T) {
	s := NewDefaultSoko()
	require.NotNil(t, s)
	assert.True(t, s.GetIsSoko())
	for i, p := range s.GetPlayers() {
		assert.True(t, p.GetSokoMode(), "player %d must evaluate with Soko ranks", i)
	}
}

func TestNewDefaultFiveCardStud_IsNotSoko(t *testing.T) {
	s := NewDefaultFiveCardStud()
	assert.False(t, s.GetIsSoko())
	for i, p := range s.GetPlayers() {
		assert.False(t, p.GetSokoMode(), "player %d must evaluate with standard ranks", i)
	}
}

// The end-to-end proof that the mode reaches the evaluator: the same five cards
// through the same production call (EvalBestHand) must rank differently on a
// Soko table and a Five Card Stud table.
func TestSoko_EvalBestHand_UsesSokoRanksOnlyInSokoMode(t *testing.T) {
	hole, door := fourFlushHand()

	soko := NewDefaultSoko()
	sp0 := soko.GetPlayers()[0]
	sp0.SetHoleCards(hole)
	sp0.SetDoorCards(door)
	assert.Equal(t, SokoHandFourFlush, sp0.EvalBestHand(), "Soko table must see the four-flush")

	plain := NewDefaultFiveCardStud()
	pp0 := plain.GetPlayers()[0]
	pp0.SetHoleCards(hole)
	pp0.SetDoorCards(door)
	assert.Equal(t, PokerHandOnePair, pp0.EvalBestHand(), "standard table must see only the pair")
}

// The Worker rebuilds the game from KV on every request. If the flag round-trips
// but is not pushed back to the players, a Soko session silently degrades into
// Five Card Stud with nothing failing anywhere.
func TestSoko_JSONRoundTrip_RestoresModeOnPlayers(t *testing.T) {
	s := NewDefaultSoko()
	require.NoError(t, s.Reset())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultFiveCardStud() // deliberately a NON-soko starting point
	require.NoError(t, json.Unmarshal(data, restored))

	assert.True(t, restored.GetIsSoko(), "the game flag must survive")
	for i, p := range restored.GetPlayers() {
		assert.True(t, p.GetSokoMode(), "player %d lost Soko mode across the round trip", i)
	}

	hole, door := fourFlushHand()
	rp := restored.GetPlayers()[0]
	rp.SetHoleCards(hole)
	rp.SetDoorCards(door)
	assert.Equal(t, SokoHandFourFlush, rp.EvalBestHand(), "the restored table must still rank Soko hands")
}

// Negative control for the round trip: a plain Five Card Stud snapshot must not
// come back as Soko (which an unconditional applySokoMode(true) would produce).
func TestSoko_JSONRoundTrip_PlainStudStaysPlain(t *testing.T) {
	s := NewDefaultFiveCardStud()
	require.NoError(t, s.Reset())
	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultSoko() // deliberately a SOKO starting point
	require.NoError(t, json.Unmarshal(data, restored))

	assert.False(t, restored.GetIsSoko())
	for i, p := range restored.GetPlayers() {
		assert.False(t, p.GetSokoMode(), "player %d wrongly kept Soko mode", i)
	}
}

// Reset must not drop the mode — it runs between every hand.
func TestSoko_ResetKeepsMode(t *testing.T) {
	s := NewDefaultSoko()
	require.NoError(t, s.Reset())
	assert.True(t, s.GetIsSoko())
	for i, p := range s.GetPlayers() {
		assert.True(t, p.GetSokoMode(), "player %d lost Soko mode across Reset", i)
	}
}

// The hand-name table has to follow the rank scale: rank 4 is Two Pair in Soko
// and a Straight in standard poker, so sharing one table would mislabel hands.
func TestSoko_HandNameFollowsTheScale(t *testing.T) {
	soko := NewDefaultSoko()
	plain := NewDefaultFiveCardStud()

	assert.Equal(t, "Two Pair", soko.getHandName(SokoHandTwoPair))
	assert.Equal(t, "Straight", plain.getHandName(PokerHandStraight))
	assert.Equal(t, SokoHandTwoPair, PokerHandStraight, "same integer, different meaning — that is why the tables differ")

	assert.Equal(t, "Four-Card Flush", soko.getHandName(SokoHandFourFlush))
	assert.Equal(t, "Four-Card Straight", soko.getHandName(SokoHandFourStraight))
}

// --- kickers -----------------------------------------------------------------
//
// The Soko rank scale and the standard scale are different integer spaces, and
// `ExtractKickers` switches on the standard one. Passing a Soko rank into it
// silently produces wrong or missing kickers, which reach the player through
// FormatKickers in both presenters. These tests drive the real showdown path
// rather than the evaluator, because that is where the two scales meet.

// sokoShowdown builds a two-seat Soko table, deals both hands, and resolves the
// showdown — the same call the production betting flow makes.
func sokoShowdown(t *testing.T, humanHole *Card, humanDoor []*Card, cpuHole *Card, cpuDoor []*Card) []FiveCardStudResult {
	t.Helper()
	cfg := DefaultSokoConfig()
	cfg.TableSize = 2
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleTAG),
	}
	s := NewSoko(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseFifthStreet)
	s.SetStartingChips([]int{1000, 1000})
	s.SetPot(200)

	players[0].SetChips(1000)
	players[0].AddHoleCard(humanHole)
	for _, c := range humanDoor {
		players[0].AddDoorCard(c)
	}
	players[1].SetChips(1000)
	players[1].AddHoleCard(cpuHole)
	for _, c := range cpuDoor {
		players[1].AddDoorCard(c)
	}

	s.SetActedFlags([]bool{true, true})
	s.resolveShowdown()
	return s.GetRoundResults()
}

func sokoResultFor(t *testing.T, results []FiveCardStudResult, idx int) FiveCardStudResult {
	t.Helper()
	for _, r := range results {
		if r.PlayerIdx == idx {
			return r
		}
	}
	t.Fatalf("no result for player %d", idx)
	return FiveCardStudResult{}
}

// Soko Two Pair is rank 4, which matches nothing in kicker.go's standard-scale
// switch — passing it straight through returns nil and the kicker vanishes.
func TestSoko_Showdown_TwoPairKeepsItsKicker(t *testing.T) {
	results := sokoShowdown(t,
		sp(5), []*Card{he(5), di(9), cl(9), sp(13)}, // two pair, K kicker
		sp(2), []*Card{he(4), di(6), cl(8), sp(10)}, // nothing
	)
	r := sokoResultFor(t, results, 0)
	assert.Equal(t, SokoHandTwoPair, r.HandRank)
	assert.Equal(t, "Two Pair", r.HandName)
	assert.Equal(t, []int{13}, r.Kickers, "the unpaired king is the kicker")
}

// Soko Three of a Kind is rank 5 — also unmatched in the standard switch.
func TestSoko_Showdown_ThreeOfAKindKeepsItsKickers(t *testing.T) {
	results := sokoShowdown(t,
		sp(5), []*Card{he(5), di(5), cl(9), sp(13)}, // trips, K+9 kickers
		sp(2), []*Card{he(4), di(6), cl(8), sp(10)},
	)
	r := sokoResultFor(t, results, 0)
	assert.Equal(t, SokoHandThreeOfAKind, r.HandRank)
	assert.Equal(t, []int{13, 9}, r.Kickers)
}

// Soko Four of a Kind is rank 9, which collides with PokerHandRoyalFlush and
// hits kicker.go's early `return nil`.
func TestSoko_Showdown_FourOfAKindKeepsItsKicker(t *testing.T) {
	results := sokoShowdown(t,
		sp(5), []*Card{he(5), di(5), cl(5), sp(13)}, // quads, K kicker
		sp(2), []*Card{he(4), di(6), cl(8), sp(10)},
	)
	r := sokoResultFor(t, results, 0)
	assert.Equal(t, SokoHandFourOfAKind, r.HandRank)
	assert.Equal(t, []int{13}, r.Kickers)
}

// The four-card hands are the reverse case: their Soko ranks (2 and 3) collide
// with TwoPair and ThreeOfAKind, so a naive pass-through invents kickers for a
// hand that has none.
func TestSoko_Showdown_FourCardHandsHaveNoKickers(t *testing.T) {
	t.Run("four-card flush", func(t *testing.T) {
		results := sokoShowdown(t,
			sp(2), []*Card{sp(5), sp(9), sp(12), he(7)},
			cl(2), []*Card{he(4), di(6), cl(8), di(10)},
		)
		r := sokoResultFor(t, results, 0)
		assert.Equal(t, SokoHandFourFlush, r.HandRank)
		assert.Equal(t, "Four-Card Flush", r.HandName)
		assert.Nil(t, r.Kickers, "a four-card flush has no pair group, so no kickers")
	})
	t.Run("four-card straight", func(t *testing.T) {
		results := sokoShowdown(t,
			sp(5), []*Card{he(6), di(7), cl(8), he(13)},
			cl(2), []*Card{he(4), di(9), cl(11), di(3)},
		)
		r := sokoResultFor(t, results, 0)
		assert.Equal(t, SokoHandFourStraight, r.HandRank)
		assert.Nil(t, r.Kickers)
	})
}

// Plain Five Card Stud must be unaffected by the kicker split.
func TestFiveCardStud_Showdown_KickersUnchangedWithoutSoko(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseFifthStreet)
	s.SetStartingChips([]int{1000, 1000})
	s.SetPot(200)
	players[0].SetChips(1000)
	players[0].AddHoleCard(sp(5))
	for _, c := range []*Card{he(5), di(9), cl(9), sp(13)} {
		players[0].AddDoorCard(c)
	}
	players[1].SetChips(1000)
	players[1].AddHoleCard(sp(2))
	for _, c := range []*Card{he(4), di(6), cl(8), sp(10)} {
		players[1].AddDoorCard(c)
	}
	s.SetActedFlags([]bool{true, true})
	s.resolveShowdown()

	r := sokoResultFor(t, s.GetRoundResults(), 0)
	assert.Equal(t, PokerHandTwoPair, r.HandRank)
	assert.Equal(t, []int{13}, r.Kickers)
}
