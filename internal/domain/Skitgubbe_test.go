//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sgCard(design, value int) *Card { return NewCard(design, value, true) }

// sgStock returns n distinct cards to keep a focused duel test inside phase one.
// An empty stock is not neutral: once it runs dry the phase ends as soon as the
// player to act has no cards, and the collected piles move into the hands.
func sgStock(n int) []*Card {
	stock := make([]*Card, 0, n)
	for i := range n {
		stock = append(stock, sgCard(CardDesignDiamond, 3+i%10))
	}
	return stock
}

func TestSkitgubbe_RankOrderPutsTheAceOnTop(t *testing.T) {
	order := []int{1, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	for i := 1; i < len(order); i++ {
		hi := sgCard(CardDesignSpade, order[i-1])
		lo := sgCard(CardDesignSpade, order[i])
		assert.Greater(t, SkitgubbeRankOrder(hi), SkitgubbeRankOrder(lo),
			"value %d must outrank value %d", order[i-1], order[i])
	}
	assert.Equal(t, 14, SkitgubbeRankOrder(sgCard(CardDesignHeart, 1)), "the ace is high")
	assert.Equal(t, 0, SkitgubbeRankOrder(nil))
}

func TestSkitgubbe_DealsThreeCardsEach(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()

	for i := range s.GetPlayers() {
		assert.Equal(t, SkitgubbeHandSize, s.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, SkitgubbeDeckSize-SkitgubbePlayerCnt*SkitgubbeHandSize, s.GetStockCount())
	assert.Equal(t, SkitgubbePhaseCollect, s.GetPhase())
	assert.Equal(t, -1, s.GetTrumpSuit(), "trump is not known until the stock runs out")
}

func TestSkitgubbe_PhaseOneIsATwoPlayerDuelWithNoSuitRule(t *testing.T) {
	// #4404 says "the OTHER players respond with the same suit or higher".
	// It is one opponent, and suit is irrelevant -- the higher card simply wins.
	s := NewDefaultSkitgubbe()
	s.Reset()
	lead := s.GetCurrentPlayerIdx()

	// Every card is legal for the leader: there is no follow-suit rule here.
	assert.Len(t, s.GetValidPlayIndices(lead), SkitgubbeHandSize)
	require.NoError(t, s.PlayCard(lead, 0))

	// Only ONE opponent answers -- the seat to the leader's left.
	assert.Equal(t, (lead+1)%SkitgubbePlayerCnt, s.GetCurrentPlayerIdx())
	assert.Len(t, s.GetValidPlayIndices(s.GetCurrentPlayerIdx()), SkitgubbeHandSize,
		"the responder may play anything too")
}

func TestSkitgubbe_TheHigherCardTakesBothAndLeadsNext(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	// Pin both hands so the duel's outcome is not a coin flip.
	s.GetPlayer(0).Reset()
	s.GetPlayer(0).AddCard(sgCard(CardDesignSpade, 5))
	s.GetPlayer(1).Reset()
	s.GetPlayer(1).AddCard(sgCard(CardDesignHeart, 9)) // different suit, higher rank
	s.SetStockForTest(sgStock(8))
	s.SetCurrentPlayerForTest(0)

	require.NoError(t, s.PlayCard(0, 0))
	require.NoError(t, s.PlayCard(1, 0))

	assert.Equal(t, 2, s.GetCollectedCount(1), "the nine wins across suits and takes both")
	assert.Zero(t, s.GetCollectedCount(0))
	assert.Equal(t, 1, s.GetDuelLeader(), "the winner leads the next duel")
}

func TestSkitgubbe_EqualRanksBounceAndTheCardsStayOnTheTable(t *testing.T) {
	// stunsa: #4404 does not mention it, and it is the phase's signature. The
	// cards are NOT taken; they pile up until someone actually wins.
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.GetPlayer(0).Reset()
	s.GetPlayer(0).AddCard(sgCard(CardDesignSpade, 7))
	s.GetPlayer(0).AddCard(sgCard(CardDesignClover, 10))
	s.GetPlayer(1).Reset()
	s.GetPlayer(1).AddCard(sgCard(CardDesignHeart, 7)) // same rank -- a bounce
	s.GetPlayer(1).AddCard(sgCard(CardDesignDiamond, 2))
	// Drawn cards are appended, so index 0 stays the pinned card throughout.
	s.SetStockForTest(sgStock(8))
	s.SetCurrentPlayerForTest(0)

	require.NoError(t, s.PlayCard(0, 0))
	require.NoError(t, s.PlayCard(1, 0))

	assert.Len(t, s.GetDuel(), 2, "the tied cards stay on the table")
	assert.Zero(t, s.GetCollectedCount(0))
	assert.Zero(t, s.GetCollectedCount(1))
	assert.Equal(t, 0, s.GetCurrentPlayerIdx(), "the same player leads again")
	assert.Equal(t, 0, s.GetDuelLeader())

	// Settle it: the ten beats the two, and takes all FOUR cards.
	require.NoError(t, s.PlayCard(0, 0))
	require.NoError(t, s.PlayCard(1, 0))
	assert.Equal(t, 4, s.GetCollectedCount(0), "the winner takes everything that piled up")
}

func TestSkitgubbe_TheLastCardDrawnSetsTrump(t *testing.T) {
	// #4404 does not mention trump at all, and phase two's whole comparison
	// depends on it.
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.GetPlayer(0).Reset()
	s.GetPlayer(0).AddCard(sgCard(CardDesignSpade, 5))
	s.GetPlayer(1).Reset()
	s.GetPlayer(1).AddCard(sgCard(CardDesignSpade, 9))
	s.SetStockForTest([]*Card{sgCard(CardDesignHeart, 4)})
	s.SetCurrentPlayerForTest(0)
	require.Equal(t, -1, s.GetTrumpSuit())

	require.NoError(t, s.PlayCard(0, 0))
	require.NoError(t, s.PlayCard(1, 0))

	assert.Equal(t, CardDesignHeart, s.GetTrumpSuit(), "the last card out of the stock names trump")
}

func TestSkitgubbe_PhaseTwoBeatsThePileAndCannotDuck(t *testing.T) {
	// #4404 says phase two is played in ASCENDING order and that a player who
	// cannot play is eliminated. It is neither: you must BEAT the pile, and if
	// you cannot you PICK IT UP.
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest([]*Card{sgCard(CardDesignSpade, 8)})
	p := s.GetPlayer(0)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 5))   // same suit, lower -- illegal
	p.AddCard(sgCard(CardDesignSpade, 10))  // same suit, higher -- legal
	p.AddCard(sgCard(CardDesignHeart, 2))   // a trump -- legal
	p.AddCard(sgCard(CardDesignClover, 13)) // another suit, no trump -- illegal
	s.SetCurrentPlayerForTest(0)

	assert.Equal(t, []int{1, 2}, s.GetValidPlayIndices(0))
	assert.ErrorContains(t, s.PlayCard(0, 0), "may not duck")
	assert.ErrorContains(t, s.PlayCard(0, 3), "may not duck")
	require.NoError(t, s.PlayCard(0, 1))
}

func TestSkitgubbe_APlayerWhoCannotBeatPicksUpRatherThanBeingEliminated(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest([]*Card{sgCard(CardDesignSpade, 13), sgCard(CardDesignHeart, 1)})
	p := s.GetPlayer(0)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 2)) // beats nothing: the top is the trump ace
	s.SetCurrentPlayerForTest(0)

	require.Empty(t, s.GetValidPlayIndices(0))
	require.NoError(t, s.PickUp(0))

	assert.Equal(t, 3, p.GetCardsSize(), "the pile joins the hand -- the player is not out")
	assert.Empty(t, s.GetPile())
	assert.False(t, s.IsFinished(0))
}

func TestSkitgubbe_PickUpIsRefusedWhenTheHandCanBeat(t *testing.T) {
	// Picking up voluntarily would be a way to duck by the back door.
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest([]*Card{sgCard(CardDesignSpade, 4)})
	p := s.GetPlayer(0)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 9))
	s.SetCurrentPlayerForTest(0)

	assert.ErrorContains(t, s.PickUp(0), "you can beat the pile")
}

func TestSkitgubbe_PickUpRejectsOtherIllegalStates(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	assert.ErrorContains(t, s.PickUp(0), "nothing to pick up", "phase one has no pile")

	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetPileForTest(nil)
	s.SetCurrentPlayerForTest(0)
	assert.ErrorContains(t, s.PickUp(0), "the pile is empty")
	assert.Error(t, s.PickUp(1), "not that player's turn")
}

func TestSkitgubbe_AnEmptyPileAcceptsAnyCard(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest(nil)
	p := s.GetPlayer(0)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 2))
	p.AddCard(sgCard(CardDesignClover, 3))
	s.SetCurrentPlayerForTest(0)

	assert.Len(t, s.GetValidPlayIndices(0), 2, "a leader may open with anything")
}

func TestSkitgubbe_CollectedCardsBecomeThePhaseTwoHand(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	require.True(t, sgPlayPhaseOne(t, s))

	assert.Equal(t, SkitgubbePhaseShed, s.GetPhase())
	assert.Zero(t, s.GetStockCount())
	assert.GreaterOrEqual(t, s.GetTrumpSuit(), 0, "trump is fixed by the time phase two starts")

	total := 0
	for i := range s.GetPlayers() {
		total += s.GetPlayer(i).GetCardsSize()
		assert.Zero(t, s.GetCollectedCount(i), "the collected pile moved into the hand")
	}
	assert.LessOrEqual(t, total, SkitgubbeDeckSize)
	assert.Positive(t, total, "somebody must have collected something")
}

func TestSkitgubbe_TheLastPlayerHoldingCardsIsTheSkitgubbe(t *testing.T) {
	for range 40 {
		s := NewDefaultSkitgubbe()
		s.Reset()
		if !sgPlayOut(t, s) {
			continue
		}
		require.True(t, s.GetGameEndFlag())

		loser := s.GetLoserIdx()
		require.GreaterOrEqual(t, loser, 0)
		assert.Positive(t, s.GetPlayer(loser).GetCardsSize(), "the loser is the one left holding cards")
		for i := range s.GetPlayers() {
			if i == loser {
				continue
			}
			assert.Zero(t, s.GetPlayer(i).GetCardsSize(), "seat %d got rid of everything", i)
		}
		return
	}
	// A skipped test verifies nothing. If 40 shuffles cannot produce one decided
	// game, the driver or the end condition is broken and that must be a failure.
	t.Fatal("no game reached a decision in 40 shuffles")
}

func TestSkitgubbe_RejectsIllegalRequests(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	cur := s.GetCurrentPlayerIdx()

	assert.Error(t, s.PlayCard(cur, -1))
	assert.Error(t, s.PlayCard(cur, 99))
	assert.Error(t, s.PlayCard((cur+1)%SkitgubbePlayerCnt, 0), "not that player's turn")
}

func TestSkitgubbe_SurvivesAKVRoundTrip(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	for range 4 {
		idx := s.GetCurrentPlayerIdx()
		if idx < 0 || s.GetPlayer(idx).GetCardsSize() == 0 {
			break
		}
		_ = s.PlayCard(idx, 0)
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultSkitgubbe()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, s.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, len(s.GetDuel()), len(restored.GetDuel()), "a bounce in progress must survive")
	assert.Equal(t, s.GetDuelLeader(), restored.GetDuelLeader())
	assert.Equal(t, s.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	for i := range s.GetPlayers() {
		assert.Equal(t, s.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "hand %d", i)
		assert.Equal(t, s.GetCollectedCount(i), restored.GetCollectedCount(i), "collected %d", i)
	}
}

func TestSkitgubbe_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("{"), NewDefaultSkitgubbe()))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), NewDefaultSkitgubbe()))

	s := NewDefaultSkitgubbe()
	s.Reset()
	data, err := json.Marshal(s)
	require.NoError(t, err)

	t.Run("invalid config", func(t *testing.T) {
		hostile := replaceJSONNumber(t, string(data), `"cd":0`, `"cd":99`)
		assert.Error(t, json.Unmarshal([]byte(hostile), NewDefaultSkitgubbe()))
	})

	t.Run("out-of-range seats are clamped", func(t *testing.T) {
		cur := fmt.Sprintf(`"ci":%d`, s.GetCurrentPlayerIdx())
		hostile := replaceJSONNumber(t, string(data), cur, `"ci":99`)
		restored := NewDefaultSkitgubbe()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, -1, restored.GetCurrentPlayerIdx())
	})

	t.Run("an unusable duel leader falls back to a real seat", func(t *testing.T) {
		// startHand indexes players[duelLeader]; -1 would panic on the next move.
		hostile := replaceJSONNumber(t, string(data), `"dl":0`, `"dl":99`)
		restored := NewDefaultSkitgubbe()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, 0, restored.GetDuelLeader())
	})
}

func TestSkitgubbe_Accessors(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()

	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(99))
	assert.Nil(t, s.GetValidPlayIndices(99))
	assert.Equal(t, 0, s.GetCollectedCount(99))
	assert.False(t, s.IsFinished(99))
	assert.Empty(t, s.GetPile())
	assert.NotEmpty(t, s.GetActionLog())
	assert.Equal(t, -1, s.GetLoserIdx())

	cfg := s.GetConfig()
	assert.NoError(t, cfg.Validate())
	s.SetConfig(cfg)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestSkitgubbe_PlayerSnapshotWithoutAnEmbeddedPlayerStillLoads(t *testing.T) {
	var p SkitgubbePlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsHuman())
}

func TestSkitgubbe_CpuNeverProducesAnIllegalMove(t *testing.T) {
	// The CPU's output goes straight back into the domain, so an off-by-one in
	// its bookkeeping shows up as a rejected move.
	for range 40 {
		s := NewDefaultSkitgubbe()
		s.Reset()
		for range 600 {
			if s.GetGameEndFlag() {
				break
			}
			idx := s.GetCurrentPlayerIdx()
			if idx < 0 {
				break
			}
			action := s.SkitgubbeCpuDecide(idx)
			if action.PickUp {
				require.NoError(t, s.PickUp(idx), "the CPU asked to pick up when it could not")
				continue
			}
			if action.HandIdx < 0 {
				break
			}
			require.NoError(t, s.PlayCard(idx, action.HandIdx),
				"the CPU proposed a move its own domain rejects")
		}
	}
}

func TestSkitgubbe_CpuBeatsWithItsCheapestWinner(t *testing.T) {
	// Beating with the ace when the nine would do throws away the stronger card.
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest([]*Card{sgCard(CardDesignSpade, 8)})
	p := s.GetPlayer(1)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 1))
	p.AddCard(sgCard(CardDesignSpade, 9))
	s.SetCurrentPlayerForTest(1)

	chosen := s.SkitgubbeCpuDecide(1).HandIdx
	assert.Equal(t, 9, p.GetCard(chosen).GetValue(), "the nine beats the eight; the ace is worth keeping")
}

func TestSkitgubbe_CpuPicksUpOnlyWhenItMust(t *testing.T) {
	s := NewDefaultSkitgubbe()
	s.Reset()
	s.SetPhaseForTest(SkitgubbePhaseShed)
	s.SetTrumpSuitForTest(CardDesignHeart)
	s.SetPileForTest([]*Card{sgCard(CardDesignHeart, 1)}) // the trump ace beats everything
	p := s.GetPlayer(1)
	p.Reset()
	p.AddCard(sgCard(CardDesignSpade, 5))
	s.SetCurrentPlayerForTest(1)

	assert.True(t, s.SkitgubbeCpuDecide(1).PickUp)

	s.SetPileForTest([]*Card{sgCard(CardDesignSpade, 2)})
	assert.False(t, s.SkitgubbeCpuDecide(1).PickUp, "with a legal card it must play")
}

// sgPlayPhaseOne drives phase one to its end with CPU decisions.
func sgPlayPhaseOne(t *testing.T, s *Skitgubbe) bool {
	t.Helper()
	for range 400 {
		if s.GetPhase() != SkitgubbePhaseCollect {
			return true
		}
		idx := s.GetCurrentPlayerIdx()
		if idx < 0 {
			return false
		}
		action := s.SkitgubbeCpuDecide(idx)
		if action.HandIdx < 0 {
			return false
		}
		require.NoError(t, s.PlayCard(idx, action.HandIdx))
	}
	return false
}

// sgPlayOut drives a whole game with CPU decisions. Returns false if it stalls.
func sgPlayOut(t *testing.T, s *Skitgubbe) bool {
	t.Helper()
	for range 2000 {
		if s.GetGameEndFlag() {
			return true
		}
		idx := s.GetCurrentPlayerIdx()
		if idx < 0 {
			return false
		}
		action := s.SkitgubbeCpuDecide(idx)
		if action.PickUp {
			if s.PickUp(idx) != nil {
				return false
			}
			continue
		}
		if action.HandIdx < 0 || s.PlayCard(idx, action.HandIdx) != nil {
			return false
		}
	}
	return false
}
