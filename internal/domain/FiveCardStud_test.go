//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fcsCard builds a *Card for tests (Ace=1, face up).
func fcsCard(d, v int) *Card {
	return NewCard(d, v, true)
}

func newTestFiveCardStud() *FiveCardStud {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	players := NewFiveCardStudPlayersForTable(cfg.TableSize)
	tc := NewTrumpCards(0)
	return NewFiveCardStud(tc, players, cfg)
}

// allCPUFiveCardStud builds a game where every seat is a CPU (no human),
// so a full game can be driven to completion without human input.
func allCPUFiveCardStud(n int) *FiveCardStud {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = n
	players := make([]*FiveCardStudPlayer, 0, n)
	styles := DefaultFiveCardStudCpuStyles(n)
	for i := 0; i < n; i++ {
		style := HoldemStyleGTO
		if i < len(styles) {
			style = styles[i]
		}
		players = append(players, NewFiveCardStudPlayer(false, style))
	}
	return NewFiveCardStud(NewTrumpCards(0), players, cfg)
}

// setupFiveCardStudForHumanAction creates a game at the given phase with human at currentTurn.
func setupFiveCardStudForHumanAction(phase int) *FiveCardStud {
	s := newTestFiveCardStud()
	s.SetPhase(phase)
	s.SetCurrentTurn(0) // human is player 0
	s.SetPot(20)
	s.SetLastBet(5)
	s.SetMinRaise(5)

	for i, p := range s.players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, i+1))
		p.AddDoorCard(fcsCard(CardDesignDiamond, i+5))
		s.SetActedFlags(make([]bool, len(s.players)))
		s.SetStartingChips([]int{1000, 1000, 1000, 1000})
	}
	return s
}

func TestNewFiveCardStud(t *testing.T) {
	s := newTestFiveCardStud()
	assert.Equal(t, FiveCardStudPhaseInit, s.GetPhase())
	assert.Equal(t, 4, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetPot())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, -1, s.GetBringInPlayerIdx())
}

func TestFiveCardStud_Reset(t *testing.T) {
	s := newTestFiveCardStud()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	err := s.Reset()
	require.NoError(t, err)

	// After reset: phase should be SecondStreet or beyond (CPUs may have acted).
	assert.GreaterOrEqual(t, s.GetPhase(), FiveCardStudPhaseSecondStreet)

	// All non-folded players should have at least 1 hole + 1 door card.
	for _, p := range s.players {
		if !p.GetFolded() {
			assert.GreaterOrEqual(t, len(p.GetHoleCards()), 1)
			assert.GreaterOrEqual(t, len(p.GetDoorCards()), 1)
		}
	}

	assert.Greater(t, s.GetPot(), 0)
	assert.GreaterOrEqual(t, s.GetBringInPlayerIdx(), 0)
}

func TestFiveCardStud_PlayerAction_Fold(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	err := s.PlayerAction(FiveCardStudActionFold, 0, 0)
	require.NoError(t, err)
}

func TestFiveCardStud_PlayerAction_Call(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	err := s.PlayerAction(FiveCardStudActionCall, 0, 100)
	require.NoError(t, err)
}

func TestFiveCardStud_PlayerAction_Check(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	s.SetLastBet(0)
	err := s.PlayerAction(FiveCardStudActionCheck, 0, 0)
	require.NoError(t, err)
}

func TestFiveCardStud_PlayerAction_Raise(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	err := s.PlayerAction(FiveCardStudActionRaise, 10, 0)
	require.NoError(t, err)
}

func TestFiveCardStud_PlayerAction_Errors(t *testing.T) {
	t.Run("game ended", func(t *testing.T) {
		s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
		s.SetGameEndFlag(true)
		assert.Error(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	})
	t.Run("wrong phase - init", func(t *testing.T) {
		s := setupFiveCardStudForHumanAction(FiveCardStudPhaseInit)
		assert.Error(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	})
	t.Run("wrong phase - showdown", func(t *testing.T) {
		s := setupFiveCardStudForHumanAction(FiveCardStudPhaseShowdown)
		assert.Error(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	})
	t.Run("not human turn", func(t *testing.T) {
		s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
		s.SetCurrentTurn(1)
		assert.Error(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	})
}

// --- bring-in (lowest door card) ---

func TestFiveCardStud_DetermineBringIn(t *testing.T) {
	s := newTestFiveCardStud()
	s.players[0].AddDoorCard(fcsCard(CardDesignSpade, 10))  // 10♠
	s.players[1].AddDoorCard(fcsCard(CardDesignHeart, 3))   // 3♥ (lowest)
	s.players[2].AddDoorCard(fcsCard(CardDesignDiamond, 7)) // 7♦
	s.players[3].AddDoorCard(fcsCard(CardDesignClover, 13)) // K♣

	assert.Equal(t, 1, s.determineBringIn())
}

func TestFiveCardStud_DetermineBringIn_TieBreakBySuit(t *testing.T) {
	s := newTestFiveCardStud()
	s.players[0].AddDoorCard(fcsCard(CardDesignSpade, 5))    // 5♠ suit rank 4
	s.players[1].AddDoorCard(fcsCard(CardDesignClover, 5))   // 5♣ suit rank 1 (lowest)
	s.players[2].AddDoorCard(fcsCard(CardDesignHeart, 5))    // 5♥ suit rank 3
	s.players[3].AddDoorCard(fcsCard(CardDesignDiamond, 10)) // 10♦

	assert.Equal(t, 1, s.determineBringIn())
}

func TestFiveCardStud_DetermineBringIn_AceIsHigh(t *testing.T) {
	s := newTestFiveCardStud()
	s.players[0].AddDoorCard(fcsCard(CardDesignSpade, 1)) // A♠ treated as high
	s.players[1].AddDoorCard(fcsCard(CardDesignHeart, 4)) // 4♥ (lowest)
	s.players[2].AddDoorCard(fcsCard(CardDesignDiamond, 9))
	s.players[3].AddDoorCard(fcsCard(CardDesignClover, 12))

	assert.Equal(t, 1, s.determineBringIn())
}

func TestFiveCardStud_DetermineBettingLeader(t *testing.T) {
	s := newTestFiveCardStud()
	s.SetPhase(FiveCardStudPhaseThirdStreet)

	// Player 0: pair of 10s showing
	s.players[0].AddDoorCard(fcsCard(CardDesignSpade, 10))
	s.players[0].AddDoorCard(fcsCard(CardDesignHeart, 10))
	// Player 1: high card
	s.players[1].AddDoorCard(fcsCard(CardDesignDiamond, 1))
	s.players[1].AddDoorCard(fcsCard(CardDesignClover, 5))
	// Player 2: high card
	s.players[2].AddDoorCard(fcsCard(CardDesignSpade, 8))
	s.players[2].AddDoorCard(fcsCard(CardDesignHeart, 7))
	// Player 3 folded
	s.players[3].SetFolded(true)

	assert.Equal(t, 0, s.determineBettingLeader())
}

// --- full betting round: bet / call / raise / fold ---

func TestFiveCardStud_BettingRound(t *testing.T) {
	// Build an all-human-style game by making CPUs not act: set all to human.
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseSecondStreet)
	for i, p := range players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, i+2))
		p.AddDoorCard(fcsCard(CardDesignDiamond, i+5))
	}
	s.SetActedFlags(make([]bool, 4))
	s.SetStartingChips([]int{1000, 1000, 1000, 1000})
	s.SetCurrentTurn(0)
	s.SetLastBet(0)

	// Player 0 bets.
	require.NoError(t, s.PlayerAction(FiveCardStudActionBet, cfg.SmallBet, 0))
	assert.Equal(t, cfg.SmallBet, s.GetLastBet())
	assert.Equal(t, 1, s.GetCurrentTurn())

	// Player 1 calls.
	require.NoError(t, s.PlayerAction(FiveCardStudActionCall, 0, 0))
	assert.Equal(t, 2, s.GetCurrentTurn())

	// Player 2 raises.
	require.NoError(t, s.PlayerAction(FiveCardStudActionRaise, cfg.SmallBet*2, 0))
	assert.Equal(t, cfg.SmallBet*2, s.GetLastBet())
	assert.Equal(t, 3, s.GetCurrentTurn())

	// Player 3 folds.
	require.NoError(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	assert.True(t, players[3].GetFolded())
}

// --- fold wins without showdown ---

func TestFiveCardStud_FoldWinsWithoutShowdown(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseSecondStreet)
	for _, p := range players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, 5))
		p.AddDoorCard(fcsCard(CardDesignDiamond, 6))
	}
	s.SetActedFlags([]bool{false, true}) // player 1 already brought in / acted
	s.SetStartingChips([]int{1000, 1000})
	s.SetPot(20)
	s.SetCurrentTurn(0)
	s.SetLastBet(2)

	require.NoError(t, s.PlayerAction(FiveCardStudActionFold, 0, 0))
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
	results := s.GetRoundResults()
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].PlayerIdx) // player 1 wins the pot
	assert.Greater(t, results[0].WonAmount, 0)
}

// --- showdown hand comparison ---

func TestFiveCardStud_Showdown_BestHandWins(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseFifthStreet)
	s.SetStartingChips([]int{1000, 1000})
	s.SetPot(100)

	// Player 0: flush (all spades)
	players[0].SetChips(1000)
	players[0].AddHoleCard(fcsCard(CardDesignSpade, 2))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 5))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 8))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 11))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 13))
	// Player 1: high card only
	players[1].SetChips(1000)
	players[1].AddHoleCard(fcsCard(CardDesignHeart, 2))
	players[1].AddDoorCard(fcsCard(CardDesignClover, 5))
	players[1].AddDoorCard(fcsCard(CardDesignDiamond, 7))
	players[1].AddDoorCard(fcsCard(CardDesignHeart, 9))
	players[1].AddDoorCard(fcsCard(CardDesignClover, 12))

	s.SetActedFlags([]bool{true, true})
	s.resolveShowdown()

	results := s.GetRoundResults()
	require.NotEmpty(t, results)
	// The flush player (0) should win the pot.
	var p0Won int
	for _, r := range results {
		if r.PlayerIdx == 0 {
			p0Won = r.WonAmount
		}
	}
	assert.Greater(t, p0Won, 0)
}

func TestFiveCardStud_IsHumanTurn(t *testing.T) {
	s := newTestFiveCardStud()
	s.SetCurrentTurn(0)
	assert.True(t, s.IsHumanTurn())
	s.SetCurrentTurn(1)
	assert.False(t, s.IsHumanTurn())
}

func TestFiveCardStud_Muck(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseShowdown)
	s.SetRoundResults([]FiveCardStudResult{
		{PlayerIdx: 0, WonAmount: 0},
		{PlayerIdx: 1, WonAmount: 50},
	})
	assert.True(t, s.IsMuckAvailable())
	require.NoError(t, s.Muck())
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
	assert.True(t, s.roundResults[0].Mucked)
}

func TestFiveCardStud_ShowHand(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseShowdown)
	require.NoError(t, s.ShowHand())
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
}

func TestFiveCardStud_Muck_WrongPhase(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	assert.Error(t, s.Muck())
}

func TestFiveCardStud_Getters(t *testing.T) {
	s := newTestFiveCardStud()
	s.SetPhase(FiveCardStudPhaseSecondStreet)
	s.SetPot(100)
	s.SetDealerIdx(2)
	s.SetCurrentTurn(1)
	s.SetLastBet(10)
	s.SetMinRaise(5)
	s.SetGameEndFlag(false)
	s.SetBringInPlayerIdx(3)

	assert.Equal(t, FiveCardStudPhaseSecondStreet, s.GetPhase())
	assert.Equal(t, 100, s.GetPot())
	assert.Equal(t, 2, s.GetDealerIdx())
	assert.Equal(t, 1, s.GetCurrentTurn())
	assert.Equal(t, 10, s.GetLastBet())
	assert.Equal(t, 5, s.GetMinRaise())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, 3, s.GetBringInPlayerIdx())
	assert.Equal(t, 4, s.GetConfig().TableSize)
	assert.NotNil(t, s.GetPlayer(0))
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(99))
}

func TestFiveCardStud_Config(t *testing.T) {
	s := newTestFiveCardStud()
	newCfg := DefaultFiveCardStudConfig()
	newCfg.Ante = 5
	s.SetConfig(newCfg)
	assert.Equal(t, 5, s.GetConfig().Ante)
}

func TestFiveCardStud_Resize(t *testing.T) {
	s := newTestFiveCardStud()
	s.Resize(NewFiveCardStudPlayersForTable(3))
	assert.Equal(t, 3, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetHandCount())
}

func TestFiveCardStud_JSON(t *testing.T) {
	s := newTestFiveCardStud()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	require.NoError(t, s.Reset())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var restored FiveCardStud
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetPot(), restored.GetPot())
	assert.Equal(t, s.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, s.GetHandCount(), restored.GetHandCount())
	assert.Equal(t, s.GetBringInPlayerIdx(), restored.GetBringInPlayerIdx())
}

func TestFiveCardStud_JSON_MaxSlice(t *testing.T) {
	badJSON := `{"pl":[` + func() string {
		s := ""
		for i := 0; i < 1001; i++ {
			if i > 0 {
				s += ","
			}
			s += `{"p":{"c":[]},"ch":{"c":0},"bp":{},"ih":false,"hc":[],"dc":[]}`
		}
		return s
	}() + `]}`
	var restored FiveCardStud
	assert.Error(t, json.Unmarshal([]byte(badJSON), &restored))
}

func TestFiveCardStud_JSON_InvalidPhase(t *testing.T) {
	var restored FiveCardStud
	// Phase out of range with an otherwise-valid default config.
	bad := `{"cf":` + mustJSON(t, DefaultFiveCardStudConfig()) + `,"ph":99}`
	assert.Error(t, json.Unmarshal([]byte(bad), &restored))
}

func TestFiveCardStud_JSON_InvalidConfig(t *testing.T) {
	var restored FiveCardStud
	bad := `{"cf":{"Ante":0},"ph":0}` // ante < 1 invalid
	assert.Error(t, json.Unmarshal([]byte(bad), &restored))
}

func TestFiveCardStud_JSON_NegativePot(t *testing.T) {
	var restored FiveCardStud
	bad := `{"cf":` + mustJSON(t, DefaultFiveCardStudConfig()) + `,"ph":0,"pt":-5}`
	assert.Error(t, json.Unmarshal([]byte(bad), &restored))
}

func mustJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func TestFiveCardStud_CurrentBetSize(t *testing.T) {
	s := newTestFiveCardStud()

	s.SetPhase(FiveCardStudPhaseSecondStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(FiveCardStudPhaseThirdStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(FiveCardStudPhaseFourthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())

	s.SetPhase(FiveCardStudPhaseFifthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())
}

func TestFiveCardStud_CPUDecide_AllStyles(t *testing.T) {
	styles := []FiveCardStudPlayStyle{
		HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO,
	}
	for _, style := range styles {
		s := newTestFiveCardStud()
		s.SetPhase(FiveCardStudPhaseSecondStreet)
		for _, p := range s.players {
			p.SetChips(1000)
			p.SetPlayStyle(style)
			p.AddHoleCard(fcsCard(CardDesignSpade, 10))
			p.AddDoorCard(fcsCard(CardDesignHeart, 10))
		}
		s.SetLastBet(5)
		action, amount := s.cpuDecide(1)
		assert.GreaterOrEqual(t, action, FiveCardStudActionFold)
		assert.GreaterOrEqual(t, amount, 0)
	}
}

func TestFiveCardStud_EvalThirdStreetStrength(t *testing.T) {
	s := newTestFiveCardStud()
	// pair gives a higher strength than a single high card
	s.players[0].AddHoleCard(fcsCard(CardDesignSpade, 13))
	s.players[0].AddDoorCard(fcsCard(CardDesignHeart, 13))
	pairStrength := s.evalThirdStreetStrength(0)

	s.players[1].AddHoleCard(fcsCard(CardDesignSpade, 2))
	s.players[1].AddDoorCard(fcsCard(CardDesignHeart, 7))
	highStrength := s.evalThirdStreetStrength(1)

	assert.Greater(t, pairStrength, highStrength)
}

func TestFiveCardStud_HandleCpuActionError(t *testing.T) {
	s := setupFiveCardStudForHumanAction(FiveCardStudPhaseSecondStreet)
	s.SetLastBet(50)
	s.players[1].SetCurrentBet(0)
	s.handleCpuActionError(1, FiveCardStudActionBet, assert.AnError)
	assert.Error(t, s.GetLastCpuError())
}

// --- multi-config full CPU game to completion ---

func TestFiveCardStud_FullCPUGameTerminates(t *testing.T) {
	for _, n := range []int{2, 4, 6} {
		n := n
		t.Run("players", func(t *testing.T) {
			s := allCPUFiveCardStud(n)
			err := s.Reset()
			require.NoError(t, err)

			// With no human seat, a single Reset must drive the hand to the end
			// (all CPU betting + showdown) without exceeding the iteration guard.
			assert.True(t, s.GetGameEndFlag(), "game should terminate for %d players", n)
			assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
			require.NoError(t, s.GetLastCpuError())
		})
	}
}

func TestFiveCardStud_FullCPUGame_MultipleHands(t *testing.T) {
	s := allCPUFiveCardStud(4)
	for i := 0; i < 10; i++ {
		require.NoError(t, s.Reset())
		assert.True(t, s.GetGameEndFlag())
	}
}

func TestFiveCardStud_Rebuy_WrongPhase(t *testing.T) {
	s := newTestFiveCardStud()
	assert.Error(t, s.Rebuy())
	assert.Error(t, s.SkipRebuy())
	assert.Error(t, s.Addon())
	assert.Error(t, s.SkipAddon())
}
