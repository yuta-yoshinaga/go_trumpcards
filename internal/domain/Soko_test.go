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
