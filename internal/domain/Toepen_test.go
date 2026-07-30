//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toepenCard(design, value int) *Card { return NewCard(design, value, true) }

func TestToepen_RankOrder(t *testing.T) {
	// THE rule of this game: 10 > 9 > 8 > 7 > A > K > Q > J. The pip cards beat
	// the ace and the court cards are the weakest -- the standard order turned
	// inside out. #4411 calls the jack "the second highest"; it is the LOWEST.
	// An earlier note of mine called it the highest, which was also wrong.
	order := []int{10, 9, 8, 7, 1, 13, 12, 11}
	for i := 1; i < len(order); i++ {
		hi := toepenCard(CardDesignSpade, order[i-1])
		lo := toepenCard(CardDesignSpade, order[i])
		assert.Greater(t, ToepenRankOrder(hi), ToepenRankOrder(lo),
			"value %d must outrank value %d", order[i-1], order[i])
	}

	assert.Equal(t, 8, ToepenRankOrder(toepenCard(CardDesignHeart, 10)), "the ten is top")
	assert.Equal(t, 1, ToepenRankOrder(toepenCard(CardDesignHeart, 11)), "the jack is bottom")
	assert.Greater(t, ToepenRankOrder(toepenCard(CardDesignHeart, 7)),
		ToepenRankOrder(toepenCard(CardDesignHeart, 1)), "even the seven beats the ace")
	assert.Equal(t, 0, ToepenRankOrder(nil))
}

func TestToepen_DeckIsThirtyTwoCards(t *testing.T) {
	deck := newToepenDeck()
	assert.Len(t, deck, ToepenDeckSize)

	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		assert.NotZero(t, ToepenRankOrder(c), "every card in the pack must have a rank")
	}
	// 2-6 are not in a 32-card pack.
	for v := 2; v <= 6; v++ {
		assert.False(t, seen[[2]int{CardDesignSpade, v}], "value %d is not in the pack", v)
	}
}

func TestToepen_Poverty(t *testing.T) {
	// A hand of nothing but the four weakest ranks may be thrown in. One pip
	// card and it is an ordinary hand.
	weak := []*Card{
		toepenCard(CardDesignSpade, 1), toepenCard(CardDesignHeart, 13),
		toepenCard(CardDesignClover, 12), toepenCard(CardDesignDiamond, 11),
	}
	assert.True(t, ToepenIsPoverty(weak))

	withPip := append([]*Card{toepenCard(CardDesignSpade, 7)}, weak[1:]...)
	assert.False(t, ToepenIsPoverty(withPip), "a seven outranks the ace, so this is not poverty")

	assert.False(t, ToepenIsPoverty(nil))
	assert.False(t, ToepenIsPoverty([]*Card{nil}))
}

func TestToepen_DealsFourCardsEach(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()

	for i := range tp.GetPlayers() {
		assert.Equal(t, ToepenHandSize, tp.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, 1, tp.GetHandNumber())
	assert.Equal(t, 1, tp.GetStake(), "a fresh hand is worth one life")
	assert.Equal(t, ToepenPhasePlay, tp.GetPhase())
}

func TestToepen_MustFollowSuit(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	lead := tp.GetCurrentPlayerIdx()

	// Before anything is led every card is legal.
	assert.Len(t, tp.GetValidPlayIndices(lead), ToepenHandSize)
	require.NoError(t, tp.PlayCard(lead, 0))

	suit := tp.GetLeadSuit()
	next := tp.GetCurrentPlayerIdx()
	p := tp.GetPlayer(next)

	var hasSuit bool
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == suit {
			hasSuit = true
			break
		}
	}
	valid := tp.GetValidPlayIndices(next)
	if hasSuit {
		for _, i := range valid {
			assert.Equal(t, suit, p.GetCard(i).GetDesign(), "holding the suit, only the suit is legal")
		}
	} else {
		assert.Len(t, valid, p.GetCardsSize(), "without the suit, anything goes")
	}
}

func TestToepen_PlayRejectsAnOffSuitCardWhenTheSuitIsHeld(t *testing.T) {
	for range 200 {
		tp := NewDefaultToepen()
		tp.Reset()
		lead := tp.GetCurrentPlayerIdx()
		require.NoError(t, tp.PlayCard(lead, 0))

		next := tp.GetCurrentPlayerIdx()
		p := tp.GetPlayer(next)
		suit := tp.GetLeadSuit()
		valid := tp.GetValidPlayIndices(next)
		if len(valid) == p.GetCardsSize() {
			continue // void in the suit; not the case under test
		}
		for i := range p.GetCardsSize() {
			if p.GetCard(i).GetDesign() != suit {
				assert.ErrorContains(t, tp.PlayCard(next, i), "follow suit")
				return
			}
		}
	}
	t.Skip("no deal produced a followable off-suit card")
}

func TestToepen_HighestOfTheLedSuitTakesTheTrick(t *testing.T) {
	// No trumps: an off-suit card can never win, however high its rank.
	tp := NewDefaultToepen()
	tp.Reset()

	before := tp.GetTrickNumber()
	for range ToepenPlayerCnt {
		idx := tp.GetCurrentPlayerIdx()
		require.NoError(t, tp.PlayCard(idx, tp.GetValidPlayIndices(idx)[0]))
	}
	assert.Equal(t, before+1, tp.GetTrickNumber())

	winner := tp.GetLastTrickWinner()
	require.GreaterOrEqual(t, winner, 0)
	assert.Equal(t, winner, tp.GetCurrentPlayerIdx(), "the winner leads next")
}

func TestToepen_ToepRaisesTheStakeAndAsksEveryoneElse(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()

	require.NoError(t, tp.Toep(0))
	assert.Equal(t, 2, tp.GetStake())
	assert.Equal(t, ToepenPhaseRespond, tp.GetPhase())
	assert.Equal(t, 0, tp.GetKnockerIdx())
	assert.Equal(t, 1, tp.GetPendingRespondent(), "the seat after the knocker answers first")

	// Cannot play a card while a response is outstanding.
	assert.ErrorContains(t, tp.PlayCard(tp.GetCurrentPlayerIdx(), 0), "toep is pending")

	for i := 1; i < ToepenPlayerCnt; i++ {
		require.NoError(t, tp.Respond(i, true))
	}
	assert.Equal(t, ToepenPhasePlay, tp.GetPhase())
	assert.Equal(t, -1, tp.GetPendingRespondent())
	assert.Equal(t, -1, tp.GetKnockerIdx())
}

func TestToepen_FoldingCostsTheStakeBeforeTheRaise(t *testing.T) {
	// Folding on the first knock costs one life, not two: you did not take on
	// the raise you are declining.
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))
	require.Equal(t, 2, tp.GetStake())

	require.NoError(t, tp.Respond(1, false))
	assert.True(t, tp.IsFolded(1))
	assert.Equal(t, 1, tp.GetLives(1), "one life for folding to the first knock")
	assert.False(t, tp.IsFolded(0))
}

func TestToepen_ASecondKnockCostsMoreToFold(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))
	for i := 1; i < ToepenPlayerCnt; i++ {
		require.NoError(t, tp.Respond(i, true))
	}
	require.NoError(t, tp.Toep(1))
	require.Equal(t, 3, tp.GetStake())

	require.NoError(t, tp.Respond(2, false))
	assert.Equal(t, 2, tp.GetLives(2), "two lives for folding to the second knock")
}

func TestToepen_OnlyTheLastTrickWinnerEscapesThePenalty(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	require.True(t, toepenPlayHand(t, tp))

	winner := tp.GetLastTrickWinner()
	require.GreaterOrEqual(t, winner, 0)
	assert.Equal(t, 0, tp.GetLives(winner), "taking the final trick costs nothing")
	for i := range tp.GetPlayers() {
		if i == winner {
			continue
		}
		assert.Equal(t, 1, tp.GetLives(i), "everyone else pays the stake")
	}
}

func TestToepen_ARaisedHandCostsMore(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))
	for i := 1; i < ToepenPlayerCnt; i++ {
		require.NoError(t, tp.Respond(i, true))
	}
	require.True(t, toepenPlayHand(t, tp))

	winner := tp.GetLastTrickWinner()
	for i := range tp.GetPlayers() {
		if i == winner {
			assert.Equal(t, 0, tp.GetLives(i))
			continue
		}
		assert.Equal(t, 2, tp.GetLives(i), "one knock means two lives")
	}
}

func TestToepen_TenLivesEliminatesAndTheLastPlayerStandingWins(t *testing.T) {
	// Redeal until seat 0 takes the final trick, rather than skipping when it
	// does not: a skipped test verifies nothing, and this is the only check
	// that elimination and the end-of-game path work at all.
	for range 400 {
		tp := NewDefaultToepen()
		tp.Reset()
		// Put the other three one life from out; settling the hand finishes them.
		for i := 1; i < ToepenPlayerCnt; i++ {
			tp.SetLives(i, ToepenMaxLives-1)
		}
		require.True(t, toepenPlayHand(t, tp))
		if tp.GetLastTrickWinner() != 0 {
			continue
		}

		for i := 1; i < ToepenPlayerCnt; i++ {
			assert.True(t, tp.IsEliminated(i), "seat %d should be out", i)
			assert.Equal(t, ToepenMaxLives, tp.GetLives(i), "lives are capped at the limit")
		}
		assert.False(t, tp.IsEliminated(0), "the last trick spared seat 0")
		assert.True(t, tp.GetGameEndFlag())
		assert.Equal(t, 0, tp.GetWinnerIdx())
		assert.Error(t, tp.NextHand(), "the game is over")
		return
	}
	t.Fatal("seat 0 never took a final trick in 400 deals -- the trick resolution is suspect")
}

func TestToepen_NextHandRejectsAMidHandCall(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	assert.Error(t, tp.NextHand())
}

func TestToepen_NextHandRotatesTheDealer(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	dealer := tp.GetDealerIdx()
	require.True(t, toepenPlayHand(t, tp))
	if tp.GetGameEndFlag() {
		t.Skip("the game ended in one hand")
	}
	require.NoError(t, tp.NextHand())
	assert.NotEqual(t, dealer, tp.GetDealerIdx())
	assert.Equal(t, 2, tp.GetHandNumber())
	assert.Equal(t, 1, tp.GetStake(), "the stake resets with the hand")
}

func TestToepen_RejectsIllegalRequests(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	cur := tp.GetCurrentPlayerIdx()

	assert.Error(t, tp.PlayCard(cur, -1))
	assert.Error(t, tp.PlayCard(cur, 99))
	assert.Error(t, tp.PlayCard((cur+1)%ToepenPlayerCnt, 0), "not that player's turn")
	assert.Error(t, tp.Respond(0, true), "no toep is pending")
	assert.Error(t, tp.Toep(99), "no such player")

	require.NoError(t, tp.Toep(0))
	assert.Error(t, tp.Toep(1), "cannot knock while a response is outstanding")
	assert.Error(t, tp.Respond(0, true), "the knocker does not answer their own knock")
}

func TestToepen_SurvivesAKVRoundTrip(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))
	require.NoError(t, tp.Respond(1, false))

	data, err := json.Marshal(tp)
	require.NoError(t, err)

	restored := NewDefaultToepen()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, tp.GetStake(), restored.GetStake())
	assert.Equal(t, tp.GetPhase(), restored.GetPhase())
	assert.Equal(t, tp.GetPendingRespondent(), restored.GetPendingRespondent())
	assert.Equal(t, tp.GetKnockerIdx(), restored.GetKnockerIdx())
	for i := range tp.GetPlayers() {
		assert.Equal(t, tp.GetLives(i), restored.GetLives(i), "lives %d", i)
		assert.Equal(t, tp.IsFolded(i), restored.IsFolded(i), "folded %d", i)
		assert.Equal(t, tp.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "hand %d", i)
	}
}

func TestToepen_UnmarshalRejectsAndRepairsHostileSnapshots(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("{"), NewDefaultToepen()))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), NewDefaultToepen()))

	tp := NewDefaultToepen()
	tp.Reset()
	data, err := json.Marshal(tp)
	require.NoError(t, err)

	t.Run("invalid config", func(t *testing.T) {
		hostile := replaceJSONNumber(t, string(data), `"cd":0`, `"cd":99`)
		assert.Error(t, json.Unmarshal([]byte(hostile), NewDefaultToepen()))
	})

	t.Run("out-of-range seats are clamped", func(t *testing.T) {
		// The current seat is the one after the dealer, not necessarily 0, so
		// read it off the game rather than assuming.
		cur := fmt.Sprintf(`"ci":%d`, tp.GetCurrentPlayerIdx())
		hostile := replaceJSONNumber(t, string(data), cur, `"ci":99`)
		restored := NewDefaultToepen()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, -1, restored.GetCurrentPlayerIdx())
	})

	t.Run("a stake below one is repaired", func(t *testing.T) {
		// A zero stake would make a settled hand cost nothing, quietly
		// disabling the only way anyone loses.
		hostile := replaceJSONNumber(t, string(data), `"sk":1`, `"sk":0`)
		restored := NewDefaultToepen()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, 1, restored.GetStake())
	})

	t.Run("short boolean slices are padded to the seat count", func(t *testing.T) {
		hostile := strings.Replace(string(data), `"fd":[false,false,false,false]`, `"fd":[]`, 1)
		require.NotEqual(t, string(data), hostile, "wire format changed; update this fixture")
		restored := NewDefaultToepen()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		for i := range restored.GetPlayers() {
			assert.False(t, restored.IsFolded(i), "seat %d must be addressable", i)
		}
	})
}

func TestToepen_Accessors(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()

	assert.Nil(t, tp.GetPlayer(-1))
	assert.Nil(t, tp.GetPlayer(99))
	assert.Nil(t, tp.GetValidPlayIndices(99))
	assert.Equal(t, 0, tp.GetLives(99))
	assert.False(t, tp.IsFolded(99))
	assert.False(t, tp.IsEliminated(99))
	assert.Equal(t, -1, tp.GetLeadSuit())
	assert.Empty(t, tp.GetCurrentTrick())
	assert.NotEmpty(t, tp.GetActionLog())
	assert.Equal(t, tp.GetLeadPlayerIdx(), tp.GetCurrentPlayerIdx())

	tp.SetLives(99, 5) // out of range: a no-op, not a panic
	assert.Equal(t, 0, tp.GetLives(0))

	cfg := tp.GetConfig()
	assert.NoError(t, cfg.Validate())
	tp.SetConfig(cfg)
	assert.Equal(t, cfg, tp.GetConfig())
}

func TestToepen_PlayerSnapshotWithoutAnEmbeddedPlayerStillLoads(t *testing.T) {
	var p ToepenPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsHuman())
}

// toepenPlayHand drives a hand to its settlement, always playing the first
// legal card. Returns false if it stalls.
func toepenPlayHand(t *testing.T, tp *Toepen) bool {
	t.Helper()
	for range 100 {
		switch tp.GetPhase() {
		case ToepenPhaseHandEnd, ToepenPhaseGameEnd:
			return true
		case ToepenPhaseRespond:
			require.NoError(t, tp.Respond(tp.GetPendingRespondent(), true))
		default:
			idx := tp.GetCurrentPlayerIdx()
			if idx < 0 {
				return false
			}
			valid := tp.GetValidPlayIndices(idx)
			if len(valid) == 0 {
				return false
			}
			require.NoError(t, tp.PlayCard(idx, valid[0]))
		}
	}
	return false
}

func TestToepen_CpuNeverProducesAnIllegalMove(t *testing.T) {
	// The CPU's own output is fed straight back into the domain, so an
	// off-by-one in its bookkeeping surfaces as a rejected move.
	for range 100 {
		tp := NewDefaultToepen()
		tp.Reset()
		for range 200 {
			phase := tp.GetPhase()
			if phase == ToepenPhaseHandEnd || phase == ToepenPhaseGameEnd {
				break
			}
			if phase == ToepenPhaseRespond {
				idx := tp.GetPendingRespondent()
				require.GreaterOrEqual(t, idx, 0)
				action := tp.ToepenCpuDecide(idx)
				require.NoError(t, tp.Respond(idx, !action.Fold))
				continue
			}
			idx := tp.GetCurrentPlayerIdx()
			require.GreaterOrEqual(t, idx, 0)
			action := tp.ToepenCpuDecide(idx)
			require.GreaterOrEqual(t, action.HandIdx, 0, "a play phase needs a card")
			require.NoError(t, tp.PlayCard(idx, action.HandIdx),
				"the CPU proposed a move its own domain rejects")
		}
		require.Contains(t, []ToepenPhase{ToepenPhaseHandEnd, ToepenPhaseGameEnd}, tp.GetPhase(),
			"a CPU-driven hand must terminate")
	}
}

func TestToepen_CpuTakesTheTrickWithItsCheapestWinner(t *testing.T) {
	// Winning with the ten when the nine would do throws away the stronger card.
	for range 300 {
		tp := NewDefaultToepen()
		tp.Reset()
		lead := tp.GetCurrentPlayerIdx()
		require.NoError(t, tp.PlayCard(lead, 0))

		idx := tp.GetCurrentPlayerIdx()
		p := tp.GetPlayer(idx)
		suit := tp.GetLeadSuit()
		best := tp.currentTrickBest()

		var winners []int
		for _, i := range tp.GetValidPlayIndices(idx) {
			c := p.GetCard(i)
			if c.GetDesign() == suit && ToepenRankOrder(c) > best {
				winners = append(winners, i)
			}
		}
		if len(winners) < 2 {
			continue // need a choice for the assertion to mean anything
		}

		chosen := tp.ToepenCpuDecide(idx).HandIdx
		cheapest := winners[0]
		for _, i := range winners {
			if ToepenRankOrder(p.GetCard(i)) < ToepenRankOrder(p.GetCard(cheapest)) {
				cheapest = i
			}
		}
		assert.Equal(t, ToepenRankOrder(p.GetCard(cheapest)), ToepenRankOrder(p.GetCard(chosen)),
			"the CPU should win with the weakest card that still wins")
		return
	}
	t.Skip("no deal offered a choice of winning cards")
}

func TestToepen_CpuFoldsOnlyWithNothingStrong(t *testing.T) {
	for range 300 {
		tp := NewDefaultToepen()
		tp.Reset()
		require.NoError(t, tp.Toep(0))
		idx := tp.GetPendingRespondent()
		action := tp.ToepenCpuDecide(idx)
		assert.Equal(t, !tp.hasStrongCard(idx), action.Fold,
			"folding must track whether anything strong is left")
		return
	}
}

func TestToepen_AFoldedSeatIsSkippedByTheHandExhaustionCheck(t *testing.T) {
	// handExhausted must ignore folded and eliminated seats: a folded player's
	// hand stays full, and counting it would keep the hand running after
	// everyone still in has played out.
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))
	require.NoError(t, tp.Respond(1, false))
	for i := 2; i < ToepenPlayerCnt; i++ {
		require.NoError(t, tp.Respond(i, true))
	}
	require.True(t, tp.IsFolded(1))
	require.Equal(t, ToepenHandSize, tp.GetPlayer(1).GetCardsSize(),
		"the folded seat keeps its cards")

	require.True(t, toepenPlayHand(t, tp))
	assert.Equal(t, ToepenPhaseHandEnd, tp.GetPhase())
}

func TestToepen_HasStrongCardIsFalseForAWeakHand(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	p := tp.GetPlayer(1)
	p.Reset()
	// Court cards and the ace are the four WEAKEST ranks in this game.
	for _, v := range []int{11, 12, 13, 1} {
		p.AddCard(toepenCard(CardDesignSpade, v))
	}
	assert.False(t, tp.hasStrongCard(1))

	p.AddCard(toepenCard(CardDesignHeart, 10))
	assert.True(t, tp.hasStrongCard(1), "a ten is the strongest card there is")
	assert.False(t, tp.hasStrongCard(99), "out of range is not a strong hand")
}

func TestToepen_WinningByForcingEveryoneOutCostsNothing(t *testing.T) {
	// Toep, everyone folds, no trick is ever played. The survivor won the hand
	// by making the raise stick -- charging them the stake would make
	// intimidating the table cost exactly as much as losing to it, which
	// removes the point of toeping at all.
	tp := NewDefaultToepen()
	tp.Reset()
	require.NoError(t, tp.Toep(0))

	for i := 1; i < ToepenPlayerCnt; i++ {
		require.NoError(t, tp.Respond(i, false))
	}
	require.Equal(t, ToepenPhaseHandEnd, tp.GetPhase(), "the hand ends when only one is left")
	require.Equal(t, -1, tp.GetLastTrickWinner(), "no trick was ever played")

	assert.Equal(t, 0, tp.GetLives(0), "the survivor pays nothing")
	for i := 1; i < ToepenPlayerCnt; i++ {
		assert.Equal(t, 1, tp.GetLives(i), "seat %d folded to the first knock", i)
	}
}

func TestToepen_RedealNeedsAPovertyHand(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	p := tp.GetPlayer(0)

	// An ordinary hand cannot ask for one.
	p.Reset()
	for _, v := range []int{10, 9, 1, 13} {
		p.AddCard(toepenCard(CardDesignSpade, v))
	}
	assert.False(t, tp.CanRedeal(0))
	assert.ErrorContains(t, tp.Redeal(0), "nothing but A, K, Q and J")

	// Nothing but the four weakest ranks can.
	p.Reset()
	for _, v := range []int{1, 13, 12, 11} {
		p.AddCard(toepenCard(CardDesignSpade, v))
	}
	assert.True(t, tp.CanRedeal(0))

	hand := tp.GetHandNumber()
	require.NoError(t, tp.Redeal(0))
	assert.Equal(t, hand, tp.GetHandNumber(), "a redeal is the same hand, not the next one")
	assert.Equal(t, ToepenHandSize, tp.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, tp.GetStake(), "the stake is unchanged")
	for i := range tp.GetPlayers() {
		assert.Equal(t, 0, tp.GetLives(i), "nobody pays for a redeal")
	}
}

func TestToepen_RedealIsRefusedOncePlayHasStarted(t *testing.T) {
	// The hand is only thrown in before anyone commits a card; afterwards the
	// information is out and a redeal would erase it.
	tp := NewDefaultToepen()
	tp.Reset()
	lead := tp.GetCurrentPlayerIdx()
	require.NoError(t, tp.PlayCard(lead, tp.GetValidPlayIndices(lead)[0]))

	p := tp.GetPlayer(0)
	p.Reset()
	for _, v := range []int{1, 13, 12, 11} {
		p.AddCard(toepenCard(CardDesignSpade, v))
	}
	assert.False(t, tp.CanRedeal(0))
	assert.ErrorContains(t, tp.Redeal(0), "before any card is played")
}

func TestToepen_RedealRejectsOtherIllegalStates(t *testing.T) {
	tp := NewDefaultToepen()
	tp.Reset()
	assert.False(t, tp.CanRedeal(99))
	assert.Error(t, tp.Redeal(99))

	require.NoError(t, tp.Toep(0))
	assert.False(t, tp.CanRedeal(0), "not while a toep is outstanding")
	assert.Error(t, tp.Redeal(0))
}

func TestToepen_ForcingEveryoneOutAfterATrickAlsoCostsNothing(t *testing.T) {
	// The same win, one trick later. My first fix read the exemption off
	// lastTrickWin whenever it was set, so once a trick had completed the
	// exemption went to THAT trick's winner -- who may since have folded --
	// and the survivor who actually forced the fold-out paid anyway.
	tp := NewDefaultToepen()
	tp.Reset()

	// Play one complete trick so lastTrickWin is a real seat.
	for range ToepenPlayerCnt {
		idx := tp.GetCurrentPlayerIdx()
		require.NoError(t, tp.PlayCard(idx, tp.GetValidPlayIndices(idx)[0]))
	}
	require.Equal(t, 1, tp.GetTrickNumber())
	trickWinner := tp.GetLastTrickWinner()
	require.GreaterOrEqual(t, trickWinner, 0)

	// Now someone toeps and everyone else folds -- including the trick winner.
	knocker := (trickWinner + 1) % ToepenPlayerCnt
	require.NoError(t, tp.Toep(knocker))
	for i := 1; i < ToepenPlayerCnt; i++ {
		seat := (knocker + i) % ToepenPlayerCnt
		require.NoError(t, tp.Respond(seat, false))
	}
	require.Equal(t, ToepenPhaseHandEnd, tp.GetPhase())

	assert.Equal(t, 0, tp.GetLives(knocker),
		"the seat left standing wins the hand, whatever happened in earlier tricks")
	for i := range tp.GetPlayers() {
		if i == knocker {
			continue
		}
		assert.Positive(t, tp.GetLives(i), "seat %d folded and pays", i)
	}
}
