//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// card is a terse constructor for the fixtures below.
func buraCard(design, value int) *Card {
	return NewCard(design, value, true)
}

func TestBura_DeckIsTheThirtySixCardShortDeck(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()

	// 3 cards each + the trump indicator + the remaining stock == 36.
	total := len(b.GetStock()) + 1 // stock excludes the trump indicator
	for i := range b.GetPlayers() {
		total += b.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 36, total, "Bura is played with a 36-card short deck")
}

func TestBura_CardPointsFollowTheAceTenTable(t *testing.T) {
	// A=11, 10=10, K=4, Q=3, J=2, everything else 0. The four suits are
	// identical, so the whole deck is worth 4*(11+10+4+3+2) == 120.
	cases := map[int]int{1: 11, 10: 10, 13: 4, 12: 3, 11: 2, 6: 0, 7: 0, 8: 0, 9: 0}
	for value, want := range cases {
		assert.Equal(t, want, BuraCardPoints(buraCard(CardDesignSpade, value)),
			"value %d", value)
	}

	total := 0
	for _, v := range ShortDeckValues {
		total += BuraCardPoints(buraCard(CardDesignHeart, v)) * 4
	}
	assert.Equal(t, BuraTotalPoints, total)
	assert.Equal(t, 120, total, "the 36-card deck holds 120 points")
}

func TestBura_RankOrderPutsTenAboveTheCourtCards(t *testing.T) {
	// A > 10 > K > Q > J > 9 > 8 > 7 > 6. The ten outranking K/Q/J is the
	// ace-ten family's defining quirk and is easy to get wrong by sorting on
	// the raw card value, where 10 sits below J=11.
	order := []int{1, 10, 13, 12, 11, 9, 8, 7, 6}
	for i := 1; i < len(order); i++ {
		hi := buraCard(CardDesignSpade, order[i-1])
		lo := buraCard(CardDesignSpade, order[i])
		assert.Greater(t, BuraRankOrder(hi), BuraRankOrder(lo),
			"value %d must outrank value %d", order[i-1], order[i])
	}
}

func TestBura_ATrumpBeatsAnyPlainCardAndAPlainSuitCannotCrossSuits(t *testing.T) {
	const trump = CardDesignHeart

	// The lowest trump beats the highest plain card.
	assert.True(t, buraCardBeats(buraCard(trump, 6), buraCard(CardDesignSpade, 1), trump))
	// ...and never the other way round.
	assert.False(t, buraCardBeats(buraCard(CardDesignSpade, 1), buraCard(trump, 6), trump))
	// Same suit compares on rank.
	assert.True(t, buraCardBeats(buraCard(CardDesignSpade, 10), buraCard(CardDesignSpade, 13), trump),
		"the ten outranks the king")
	// A different plain suit beats nothing -- it is a discard, not a response.
	assert.False(t, buraCardBeats(buraCard(CardDesignClover, 1), buraCard(CardDesignSpade, 6), trump))
}

func TestBura_BeatingACombinationNeedsEveryCardCovered(t *testing.T) {
	const trump = CardDesignHeart
	lead := []*Card{buraCard(CardDesignSpade, 13), buraCard(CardDesignSpade, 12)} // K, Q

	t.Run("every card individually beaten", func(t *testing.T) {
		resp := []*Card{buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 10)} // A, 10
		assert.True(t, buraBeatsCombination(resp, lead, trump))
	})

	t.Run("one card short is not enough", func(t *testing.T) {
		// The ace covers the king, but the jack cannot cover the queen.
		resp := []*Card{buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 11)}
		assert.False(t, buraBeatsCombination(resp, lead, trump))
	})

	t.Run("trumps cover a plain-suit combination", func(t *testing.T) {
		resp := []*Card{buraCard(trump, 6), buraCard(trump, 7)}
		assert.True(t, buraBeatsCombination(resp, lead, trump))
	})

	t.Run("a mixed response is legal when the matching works out", func(t *testing.T) {
		// The trump six covers the king; the ace of the led suit covers the queen.
		resp := []*Card{buraCard(trump, 6), buraCard(CardDesignSpade, 1)}
		assert.True(t, buraBeatsCombination(resp, lead, trump))
	})

	t.Run("greedy matching must not waste a card that only one lead card needs", func(t *testing.T) {
		// Lead A, 6 of the led suit. The response holds one trump and the led
		// suit's 7. Assigning the trump to the ace and the seven to the six
		// works; assigning the trump to the six strands the ace. A matcher that
		// pairs cards in dealt order rather than searching would reject this.
		lead := []*Card{buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 6)}
		resp := []*Card{buraCard(CardDesignSpade, 7), buraCard(trump, 6)}
		assert.True(t, buraBeatsCombination(resp, lead, trump))
	})

	t.Run("card counts must match", func(t *testing.T) {
		resp := []*Card{buraCard(trump, 6)}
		assert.False(t, buraBeatsCombination(resp, lead, trump))
	})
}

func TestBura_LeadMustBeASingleSuit(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()

	assert.NoError(t, buraValidateLead([]*Card{
		buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 6),
	}))
	assert.Error(t, buraValidateLead([]*Card{
		buraCard(CardDesignSpade, 1), buraCard(CardDesignHeart, 6),
	}), "a lead mixing suits is illegal")
	assert.Error(t, buraValidateLead(nil), "leading nothing is illegal")
	assert.Error(t, buraValidateLead([]*Card{
		buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 6),
		buraCard(CardDesignSpade, 7), buraCard(CardDesignSpade, 8),
	}), "a hand holds three cards, so four cannot be led")
}

func TestBura_InstantWinCombinations(t *testing.T) {
	const trump = CardDesignHeart

	cases := []struct {
		name string
		hand []*Card
		want BuraCombination
	}{
		{
			name: "bura is three trumps",
			hand: []*Card{buraCard(trump, 1), buraCard(trump, 6), buraCard(trump, 13)},
			want: BuraCombinationBura,
		},
		{
			name: "moscow is three aces",
			hand: []*Card{buraCard(CardDesignSpade, 1), buraCard(CardDesignClover, 1), buraCard(CardDesignDiamond, 1)},
			want: BuraCombinationMoscow,
		},
		{
			name: "little moscow is three sixes including the trump six",
			hand: []*Card{buraCard(trump, 6), buraCard(CardDesignSpade, 6), buraCard(CardDesignClover, 6)},
			want: BuraCombinationLittleMoscow,
		},
		{
			name: "three sixes without the trump six is not little moscow",
			hand: []*Card{buraCard(CardDesignSpade, 6), buraCard(CardDesignClover, 6), buraCard(CardDesignDiamond, 6)},
			want: BuraCombinationNone,
		},
		{
			name: "molodka is three cards of one plain suit",
			hand: []*Card{buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 6), buraCard(CardDesignSpade, 13)},
			want: BuraCombinationMolodka,
		},
		{
			name: "three trumps is bura, not molodka",
			hand: []*Card{buraCard(trump, 7), buraCard(trump, 8), buraCard(trump, 9)},
			want: BuraCombinationBura,
		},
		{
			name: "an ordinary hand has nothing",
			hand: []*Card{buraCard(CardDesignSpade, 7), buraCard(CardDesignClover, 8), buraCard(trump, 9)},
			want: BuraCombinationNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, BuraDetectCombination(tc.hand, trump))
		})
	}
}

func TestBura_ClaimingThirtyOneWinsAndClaimingShortLosesTheRound(t *testing.T) {
	t.Run("a true claim wins", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		b.SetPlayerPoints(0, BuraWinThreshold)

		require.NoError(t, b.Claim(0))
		assert.True(t, b.GetGameEndFlag())
		assert.Equal(t, 0, b.GetWinnerIdx())
	})

	t.Run("a false claim hands the round to the opponent", func(t *testing.T) {
		// Bura's real penalty is doubling the pot; with no pot to double, the
		// round is forfeited. Either way a wrong claim must cost something --
		// if it were free the whole claiming mechanic would be noise.
		b := NewDefaultBura()
		b.Reset()
		b.SetPlayerPoints(0, BuraWinThreshold-1)

		require.NoError(t, b.Claim(0))
		assert.True(t, b.GetGameEndFlag())
		assert.Equal(t, 1, b.GetWinnerIdx(), "the opponent takes the round")
	})
}

func TestBura_ReachingThirtyOneDoesNotWinOnItsOwn(t *testing.T) {
	// Points alone never end the round -- Bura is claim-driven, and a player
	// who sails past 31 without noticing keeps playing. Automating this would
	// delete the game's central tension.
	b := NewDefaultBura()
	b.Reset()
	b.SetPlayerPoints(0, BuraTotalPoints)

	assert.False(t, b.GetGameEndFlag())
	assert.Equal(t, -1, b.GetWinnerIdx())
}

func TestBura_HandsRefillToThreeWhileStockLasts(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()

	for i := range b.GetPlayers() {
		assert.Equal(t, BuraHandSize, b.GetPlayer(i).GetCardsSize())
	}

	// Drive the round to exhaustion; hands must stay at three until the stock
	// runs dry, then shrink monotonically -- never refill from nothing.
	seenShrink := false
	for range 200 {
		if b.GetGameEndFlag() {
			break
		}
		before := b.GetPlayer(0).GetCardsSize()
		if !buraDriveOneTrick(t, b) {
			break
		}
		after := b.GetPlayer(0).GetCardsSize()
		if len(b.GetStock()) > 0 && !seenShrink {
			assert.Equal(t, BuraHandSize, after, "hands refill while the stock lasts")
		}
		if after < before {
			seenShrink = true
		}
	}
}

func TestBura_RoundEndsWithoutAWinnerWhenNobodyClaims(t *testing.T) {
	// Real Bura throws the cards in and redeals when the hands empty with no
	// claim. The issue proposed settling on points instead, which would make
	// the 31-point target decorative.
	b := NewDefaultBura()
	b.Reset()

	for range 200 {
		if b.GetGameEndFlag() {
			break
		}
		if !buraDriveOneTrick(t, b) {
			break
		}
	}

	require.True(t, b.GetGameEndFlag(), "the round must terminate")
	assert.Equal(t, -1, b.GetWinnerIdx(), "no claim means no winner")
	assert.True(t, b.IsDraw())
}

func TestBura_TrickWinnerTakesThePointsOnTheTable(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()

	before := 0
	for i := range b.GetPlayers() {
		before += b.GetPlayerPoints(i)
	}
	require.Equal(t, 0, before)

	require.True(t, buraDriveOneTrick(t, b))

	after := 0
	for i := range b.GetPlayers() {
		after += b.GetPlayerPoints(i)
	}
	assert.GreaterOrEqual(t, after, 0)
	assert.LessOrEqual(t, after, BuraTotalPoints)
}

func TestBura_SurvivesAKVRoundTrip(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()
	require.True(t, buraDriveOneTrick(t, b))

	data, err := json.Marshal(b)
	require.NoError(t, err)

	restored := NewDefaultBura()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, b.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, len(b.GetStock()), len(restored.GetStock()))
	for i := range b.GetPlayers() {
		assert.Equal(t, b.GetPlayerPoints(i), restored.GetPlayerPoints(i), "player %d points", i)
		assert.Equal(t, b.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "player %d hand", i)
	}
}

// buraDriveOneTrick plays one full trick with whatever the current player
// holds, leading a single card and answering with a legal response when one
// exists. It returns false once the game can no longer progress.
func buraDriveOneTrick(t *testing.T, b *Bura) bool {
	t.Helper()
	if b.GetGameEndFlag() {
		return false
	}
	guard := 0
	for !b.GetGameEndFlag() {
		guard++
		require.Less(t, guard, 50, "a trick must not loop forever")

		idx := b.GetCurrentPlayerIdx()
		if idx < 0 {
			return false
		}
		p := b.GetPlayer(idx)
		if p.GetCardsSize() == 0 {
			return false
		}
		n := 1
		if lead := b.GetCurrentLead(); len(lead) > 0 {
			n = len(lead)
		}
		if p.GetCardsSize() < n {
			return false
		}
		indices := make([]int, n)
		for i := range indices {
			indices[i] = i
		}
		if err := b.PlayCards(idx, indices); err != nil {
			// A lead of several cards may be illegal (mixed suits); fall back
			// to a single card, which is always legal for the leader.
			if len(b.GetCurrentLead()) > 0 {
				return false
			}
			if err := b.PlayCards(idx, []int{0}); err != nil {
				return false
			}
		}
		if len(b.GetCurrentLead()) == 0 {
			return true // the trick resolved
		}
	}
	return false
}

// buraSetHand replaces idx's hand with exactly the given cards.
func buraSetHand(b *Bura, idx int, cards ...*Card) {
	p := b.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestBura_CpuDeclaresAWinningHandBeforeAnythingElse(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()
	b.SetTrumpSuit(CardDesignHeart)
	buraSetHand(b, 1,
		buraCard(CardDesignHeart, 1), buraCard(CardDesignHeart, 6), buraCard(CardDesignHeart, 13))

	action := b.BuraCpuDecide(1)
	assert.True(t, action.Declare, "a bura in hand ends the round immediately")
	assert.False(t, action.Claim)
}

func TestBura_CpuClaimsOnlyWhenItActuallyHasThirtyOne(t *testing.T) {
	// A CPU that claims short would forfeit, so the threshold check has to be
	// the real one -- an off-by-one here loses the CPU every round it reaches 30.
	b := NewDefaultBura()
	b.Reset()
	b.SetTrumpSuit(CardDesignHeart)
	buraSetHand(b, 1,
		buraCard(CardDesignSpade, 7), buraCard(CardDesignClover, 8), buraCard(CardDesignDiamond, 9))

	b.SetPlayerPoints(1, BuraWinThreshold-1)
	assert.False(t, b.BuraCpuDecide(1).Claim, "30 points is not 31")

	b.SetPlayerPoints(1, BuraWinThreshold)
	assert.True(t, b.BuraCpuDecide(1).Claim)
}

func TestBura_CpuLeadsItsBiggestPlainSuitGroup(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()
	b.SetTrumpSuit(CardDesignHeart)
	// Two spades worth 21 points together, plus an unrelated club.
	buraSetHand(b, 1,
		buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 10), buraCard(CardDesignClover, 7))
	b.SetCurrentPlayerIdx(1)

	action := b.BuraCpuDecide(1)
	assert.Len(t, action.Indices, 2, "leading both spades at once collects both cards' points")
	for _, i := range action.Indices {
		assert.Equal(t, CardDesignSpade, b.GetPlayer(1).GetCard(i).GetDesign())
	}
}

func TestBura_CpuAnswersWithTheCheapestCoverAndDiscardsWhenItCannotCover(t *testing.T) {
	const trump = CardDesignHeart

	t.Run("covers with the cheapest sufficient card", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		b.SetTrumpSuit(trump)
		b.SetCurrentLead([]*Card{buraCard(CardDesignSpade, 12)}) // Q
		// King covers the queen; the ace also would, but is worth keeping.
		buraSetHand(b, 1,
			buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 13), buraCard(trump, 6))
		b.SetCurrentPlayerIdx(1)

		action := b.BuraCpuDecide(1)
		require.Len(t, action.Indices, 1)
		played := b.GetPlayer(1).GetCard(action.Indices[0])
		assert.Equal(t, 13, played.GetValue(), "the king is the cheapest card that still beats the queen")
	})

	t.Run("discards its lowest card when nothing covers", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		b.SetTrumpSuit(trump)
		b.SetCurrentLead([]*Card{buraCard(CardDesignSpade, 1)}) // A of a suit it cannot touch
		buraSetHand(b, 1,
			buraCard(CardDesignClover, 10), buraCard(CardDesignDiamond, 7), buraCard(CardDesignClover, 13))
		b.SetCurrentPlayerIdx(1)

		action := b.BuraCpuDecide(1)
		require.Len(t, action.Indices, 1)
		played := b.GetPlayer(1).GetCard(action.Indices[0])
		assert.Equal(t, 0, BuraCardPoints(played),
			"with no cover available the CPU must throw away a card worth nothing, not the ten")
	})
}

func TestBura_CpuNeverProducesAnIllegalMove(t *testing.T) {
	// The CPU's own output is fed straight back into PlayCards, so an
	// off-by-one in its index bookkeeping surfaces as a rejected move rather
	// than a wrong one. Drive whole games on both seats to shake that out.
	for range 100 {
		b := NewDefaultBura()
		b.Reset()
		for range 400 {
			if b.GetGameEndFlag() {
				break
			}
			idx := b.GetCurrentPlayerIdx()
			require.GreaterOrEqual(t, idx, 0)
			action := b.BuraCpuDecide(idx)
			switch {
			case action.Declare:
				require.NoError(t, b.DeclareCombination(idx))
			case action.Claim:
				require.NoError(t, b.Claim(idx))
			default:
				require.NotEmpty(t, action.Indices, "the CPU must choose something")
				require.NoError(t, b.PlayCards(idx, action.Indices),
					"the CPU proposed a move its own domain rejects")
			}
		}
		require.True(t, b.GetGameEndFlag(), "a CPU-vs-CPU round must terminate")
	}
}

func TestBura_PlayCardsRejectsIllegalRequests(t *testing.T) {
	newGame := func() *Bura {
		b := NewDefaultBura()
		b.Reset()
		b.SetCurrentPlayerIdx(0)
		return b
	}

	t.Run("not your turn", func(t *testing.T) {
		b := newGame()
		assert.Error(t, b.PlayCards(1, []int{0}))
	})

	t.Run("no such player", func(t *testing.T) {
		b := newGame()
		b.SetCurrentPlayerIdx(9)
		assert.Error(t, b.PlayCards(9, []int{0}))
	})

	t.Run("index out of range", func(t *testing.T) {
		b := newGame()
		assert.Error(t, b.PlayCards(0, []int{99}))
		assert.Error(t, b.PlayCards(0, []int{-1}))
	})

	t.Run("the same card twice", func(t *testing.T) {
		// Without this check a player could turn one card into a two-card
		// lead, which both doubles its points and outnumbers the responder.
		b := newGame()
		assert.Error(t, b.PlayCards(0, []int{0, 0}))
	})

	t.Run("empty selection", func(t *testing.T) {
		b := newGame()
		assert.Error(t, b.PlayCards(0, []int{}))
	})

	t.Run("a lead of mixed suits", func(t *testing.T) {
		b := newGame()
		b.SetTrumpSuit(CardDesignHeart)
		buraSetHand(b, 0,
			buraCard(CardDesignSpade, 1), buraCard(CardDesignClover, 6), buraCard(CardDesignSpade, 7))
		assert.Error(t, b.PlayCards(0, []int{0, 1}))
		assert.NoError(t, b.PlayCards(0, []int{0, 2}), "two spades are a legal lead")
	})

	t.Run("a response of the wrong size", func(t *testing.T) {
		b := newGame()
		b.SetTrumpSuit(CardDesignHeart)
		buraSetHand(b, 0, buraCard(CardDesignSpade, 1), buraCard(CardDesignSpade, 7))
		require.NoError(t, b.PlayCards(0, []int{0, 1}))
		require.Equal(t, 1, b.GetCurrentPlayerIdx())
		assert.Error(t, b.PlayCards(1, []int{0}), "the lead was two cards")
	})

	t.Run("after the round is over", func(t *testing.T) {
		b := newGame()
		require.NoError(t, b.Claim(0))
		assert.Error(t, b.PlayCards(0, []int{0}))
	})
}

func TestBura_ClaimAndDeclareRejectIllegalRequests(t *testing.T) {
	t.Run("claim by a player who does not exist", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		assert.Error(t, b.Claim(9))
	})

	t.Run("claim after the round is over", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		require.NoError(t, b.Claim(0))
		assert.Error(t, b.Claim(1))
	})

	t.Run("declaring without a combination leaves the round running", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		b.SetTrumpSuit(CardDesignHeart)
		buraSetHand(b, 0,
			buraCard(CardDesignSpade, 7), buraCard(CardDesignClover, 8), buraCard(CardDesignDiamond, 9))
		assert.Error(t, b.DeclareCombination(0))
		assert.False(t, b.GetGameEndFlag(), "a rejected declaration must not end the round")
		assert.Equal(t, -1, b.GetWinnerIdx())
	})

	t.Run("declaring a real combination wins", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		b.SetTrumpSuit(CardDesignHeart)
		buraSetHand(b, 0,
			buraCard(CardDesignHeart, 1), buraCard(CardDesignHeart, 7), buraCard(CardDesignHeart, 8))
		require.NoError(t, b.DeclareCombination(0))
		assert.True(t, b.GetGameEndFlag())
		assert.Equal(t, 0, b.GetWinnerIdx())
	})

	t.Run("declare by a player who does not exist", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		assert.Error(t, b.DeclareCombination(9))
	})

	t.Run("declare after the round is over", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		require.NoError(t, b.Claim(0))
		assert.Error(t, b.DeclareCombination(0))
	})
}

func TestBura_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	// Workers rebuild the game from KV on every request, so this runs on bytes
	// the process did not write.
	t.Run("malformed JSON", func(t *testing.T) {
		assert.Error(t, json.Unmarshal([]byte("{"), NewDefaultBura()))
	})

	t.Run("no players", func(t *testing.T) {
		assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), NewDefaultBura()))
	})

	t.Run("invalid config", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		data, err := json.Marshal(b)
		require.NoError(t, err)
		hostile := replaceJSONNumber(t, string(data), `"cd":0`, `"cd":99`)
		assert.Error(t, json.Unmarshal([]byte(hostile), NewDefaultBura()))
	})

	t.Run("out-of-range player indices are clamped, not trusted", func(t *testing.T) {
		b := NewDefaultBura()
		b.Reset()
		data, err := json.Marshal(b)
		require.NoError(t, err)
		hostile := replaceJSONNumber(t, string(data), `"cp":0`, `"cp":99`)

		restored := NewDefaultBura()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, -1, restored.GetCurrentPlayerIdx(),
			"an index past the players slice must not survive into play")
	})
}

func TestBura_AccessorsReportTheDealtState(t *testing.T) {
	b := NewDefaultBura()
	b.Reset()

	assert.Equal(t, BuraPhasePlay, b.GetPhase())
	assert.Equal(t, 0, b.GetTrickNumber())
	assert.Equal(t, 0, b.GetLeadPlayerIdx())
	assert.NotNil(t, b.GetTrumpCard(), "the trump indicator is face up until it is drawn")
	assert.Equal(t, b.GetTrumpCard().GetDesign(), b.GetTrumpSuit())
	assert.NotEmpty(t, b.GetActionLog(), "the deal is logged")
	assert.Nil(t, b.GetPlayer(-1))
	assert.Nil(t, b.GetPlayer(99))
	assert.Equal(t, 0, b.GetPlayerPoints(99))

	b.SetPlayerPoints(99, 10) // out of range: must be a no-op, not a panic
	assert.Equal(t, 0, b.GetPlayerPoints(0))

	cfg := b.GetConfig()
	assert.NoError(t, cfg.Validate())
	b.SetConfig(cfg)
	assert.Equal(t, cfg, b.GetConfig())

	assert.Equal(t, 0, BuraCardPoints(nil))
	assert.Equal(t, 0, BuraRankOrder(nil))
	assert.Equal(t, BuraCombinationNone, BuraDetectCombination(nil, CardDesignHeart))
	assert.Equal(t, BuraCombinationNone,
		BuraDetectCombination([]*Card{nil, nil, nil}, CardDesignHeart))
	assert.False(t, buraCardBeats(nil, buraCard(CardDesignSpade, 1), CardDesignHeart))
}

// replaceJSONNumber swaps one literal field assignment in a JSON string,
// failing the test if it is not present (so a wire-format rename surfaces here
// instead of silently making the case vacuous).
func replaceJSONNumber(t *testing.T, data, old, replacement string) string {
	t.Helper()
	require.Contains(t, data, old, "wire format changed; update this fixture")
	return strings.Replace(data, old, replacement, 1)
}

func TestBura_PlayerSnapshotWithoutAnEmbeddedPlayerStillLoads(t *testing.T) {
	// A snapshot missing the embedded GamePlayer must produce a usable player
	// rather than a nil dereference on the next hand access -- Workers restore
	// from KV bytes they did not write.
	var p BuraPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsHuman())
}
