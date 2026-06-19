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

	// Player 2 raises by one small bet (the amount is the raise increment, added
	// on top of calling the outstanding 5), so the outstanding bet becomes 10.
	require.NoError(t, s.PlayerAction(FiveCardStudActionRaise, cfg.SmallBet, 0))
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

// --- Config.Validate / defaults ---

func TestFiveCardStudConfig_Defaults(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	assert.Equal(t, 1, cfg.Ante)
	assert.Equal(t, 2, cfg.BringIn)
	assert.Equal(t, 5, cfg.SmallBet)
	assert.Equal(t, 10, cfg.BigBet)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, FiveCardStudTableSize6, cfg.TableSize)
	assert.Equal(t, BettingLimitFixed, cfg.BettingLimit)
	require.NoError(t, cfg.Validate())
}

func TestFiveCardStudConfig_Validate(t *testing.T) {
	valid := DefaultFiveCardStudConfig()

	tests := []struct {
		name    string
		mutate  func(c *FiveCardStudConfig)
		wantErr bool
	}{
		{"valid default", func(c *FiveCardStudConfig) {}, false},
		{"table size 0 allowed (no change)", func(c *FiveCardStudConfig) { c.TableSize = 0 }, false},
		{"bad betting limit low", func(c *FiveCardStudConfig) { c.BettingLimit = BettingLimitType(-1) }, true},
		{"bad betting limit high", func(c *FiveCardStudConfig) { c.BettingLimit = BettingLimitType(99) }, true},
		{"ante zero", func(c *FiveCardStudConfig) { c.Ante = 0 }, true},
		{"ante negative", func(c *FiveCardStudConfig) { c.Ante = -5 }, true},
		{"bring-in zero", func(c *FiveCardStudConfig) { c.BringIn = 0 }, true},
		{"small bet zero", func(c *FiveCardStudConfig) { c.SmallBet = 0 }, true},
		{"big bet zero", func(c *FiveCardStudConfig) { c.BigBet = 0 }, true},
		{"small bet > big bet", func(c *FiveCardStudConfig) { c.SmallBet = 20; c.BigBet = 10 }, true},
		{"ante level hands zero", func(c *FiveCardStudConfig) { c.AnteLevelHands = 0 }, true},
		{"init chips zero", func(c *FiveCardStudConfig) { c.InitChips = 0 }, true},
		{"table size 1 invalid", func(c *FiveCardStudConfig) { c.TableSize = 1 }, true},
		{"table size 7 invalid", func(c *FiveCardStudConfig) { c.TableSize = 7 }, true},
		{"table size 2 valid", func(c *FiveCardStudConfig) { c.TableSize = 2 }, false},
		{"table size 6 valid", func(c *FiveCardStudConfig) { c.TableSize = 6 }, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			tc.mutate(&cfg)
			err := cfg.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidFiveCardStudTableSize(t *testing.T) {
	assert.False(t, IsValidFiveCardStudTableSize(1))
	assert.True(t, IsValidFiveCardStudTableSize(2))
	assert.True(t, IsValidFiveCardStudTableSize(6))
	assert.False(t, IsValidFiveCardStudTableSize(7))
}

func TestDefaultFiveCardStudCpuStyles(t *testing.T) {
	assert.Len(t, DefaultFiveCardStudCpuStyles(3), 2)
	assert.Len(t, DefaultFiveCardStudCpuStyles(4), 3)
	assert.Len(t, DefaultFiveCardStudCpuStyles(5), 4)
	assert.Len(t, DefaultFiveCardStudCpuStyles(6), 5)
	assert.Len(t, DefaultFiveCardStudCpuStyles(2), 1) // <=2 -> single GTO
	assert.Equal(t, fiveCardStudStyles6, DefaultFiveCardStudCpuStyles(8))
}

func TestNewFiveCardStudPlayersForTable_InvalidFallsBack(t *testing.T) {
	// Invalid size falls back to 6-max (1 human + 5 CPU).
	players := NewFiveCardStudPlayersForTable(1)
	assert.Len(t, players, FiveCardStudTableSize6+1)
	assert.True(t, players[0].GetIsHuman())
}

func TestNewDefaultFiveCardStud(t *testing.T) {
	s := NewDefaultFiveCardStud()
	assert.Equal(t, FiveCardStudPhaseInit, s.GetPhase())
	assert.Equal(t, FiveCardStudTableSize6+1, s.GetPlayerCnt())
}

// --- Rebuy / Add-on flows ---

// rebuyEnabledFiveCardStud returns a 2-handed game (human + CPU) with rebuy enabled
// and the human seat already at 0 chips, sitting in the Rebuy phase.
func rebuyEnabledFiveCardStud(t *testing.T) *FiveCardStud {
	t.Helper()
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.RebuyEnabled = true
	cfg.RebuyChips = 500
	cfg.RebuyMaxCount = 2
	cfg.AddonEnabled = false
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),  // human, broke
		NewFiveCardStudPlayer(false, HoldemStyleGTO), // cpu
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	players[0].SetChips(0)
	players[1].SetChips(1000)
	s.SetHandCount(1)
	s.SetPhase(FiveCardStudPhaseRebuy)
	s.SetRebuyPhaseType(FiveCardStudRebuyPhaseRebuy)
	s.SetRebuyCounts([]int{0, 0})
	s.SetAddonUsed([]bool{false, false})
	return s
}

func TestFiveCardStud_Rebuy_AddsChips(t *testing.T) {
	s := rebuyEnabledFiveCardStud(t)
	require.NoError(t, s.Rebuy())
	assert.Equal(t, 500, s.GetPlayer(0).GetChips())
	assert.Equal(t, 1, s.GetRebuyCounts()[0])
	assert.Equal(t, FiveCardStudRebuyPhaseNone, s.GetRebuyPhaseType())
	// continueReset dealt cards -> game advanced beyond Rebuy.
	assert.NotEqual(t, FiveCardStudPhaseRebuy, s.GetPhase())
}

func TestFiveCardStud_SkipRebuy_EndsWhenHumanBroke(t *testing.T) {
	s := rebuyEnabledFiveCardStud(t)
	require.NoError(t, s.SkipRebuy())
	// Human declined and is broke -> game over.
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
	assert.True(t, s.GetGameEndFlag())
}

func TestFiveCardStud_SkipRebuy_ContinuesWhenHumanHasChips(t *testing.T) {
	s := rebuyEnabledFiveCardStud(t)
	s.GetPlayer(0).SetChips(100) // human is no longer broke
	require.NoError(t, s.SkipRebuy())
	assert.NotEqual(t, FiveCardStudPhaseEnd, s.GetPhase())
	assert.False(t, s.GetGameEndFlag())
}

func TestFiveCardStud_Addon_AddsChips(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.AddonEnabled = true
	cfg.AddonChips = 750
	cfg.AddonAfterHand = 1
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	players[0].SetChips(200)
	players[1].SetChips(1000)
	s.SetHandCount(1)
	s.SetPhase(FiveCardStudPhaseRebuy)
	s.SetRebuyPhaseType(FiveCardStudRebuyPhaseAddon)
	s.SetAddonUsed([]bool{false, false})

	require.NoError(t, s.Addon())
	assert.Equal(t, 950, s.GetPlayer(0).GetChips()) // 200 + 750
	assert.True(t, s.GetAddonUsed()[0])
	assert.Equal(t, FiveCardStudRebuyPhaseNone, s.GetRebuyPhaseType())
}

func TestFiveCardStud_SkipAddon_Continues(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	players[0].SetChips(500)
	players[1].SetChips(1000)
	s.SetHandCount(1)
	s.SetPhase(FiveCardStudPhaseRebuy)
	s.SetRebuyPhaseType(FiveCardStudRebuyPhaseAddon)
	s.SetAddonUsed([]bool{false, false})

	require.NoError(t, s.SkipAddon())
	assert.False(t, s.GetAddonUsed()[0])
	assert.Equal(t, FiveCardStudRebuyPhaseNone, s.GetRebuyPhaseType())
	assert.NotEqual(t, FiveCardStudPhaseRebuy, s.GetPhase())
}

func TestFiveCardStud_Rebuy_WrongPhaseType(t *testing.T) {
	s := newTestFiveCardStud()
	s.SetPhase(FiveCardStudPhaseRebuy)
	// In Rebuy phase but wrong sub-type for each action.
	s.SetRebuyPhaseType(FiveCardStudRebuyPhaseAddon)
	assert.Error(t, s.Rebuy())
	assert.Error(t, s.SkipRebuy())
	s.SetRebuyPhaseType(FiveCardStudRebuyPhaseRebuy)
	assert.Error(t, s.Addon())
	assert.Error(t, s.SkipAddon())
}

func TestFiveCardStud_IsRebuyAvailable(t *testing.T) {
	s := rebuyEnabledFiveCardStud(t)
	assert.True(t, s.IsRebuyAvailable())

	// Past the rebuy period -> unavailable.
	s.SetHandCount(s.GetConfig().RebuyPeriodHands + 1)
	assert.False(t, s.IsRebuyAvailable())

	// Within period but human has chips -> unavailable.
	s.SetHandCount(1)
	s.GetPlayer(0).SetChips(100)
	assert.False(t, s.IsRebuyAvailable())

	// Rebuy disabled -> unavailable.
	cfg := s.GetConfig()
	cfg.RebuyEnabled = false
	s.SetConfig(cfg)
	s.GetPlayer(0).SetChips(0)
	assert.False(t, s.IsRebuyAvailable())

	// Re-enable but rebuy count maxed out -> unavailable.
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 1
	s.SetConfig(cfg)
	s.SetRebuyCounts([]int{1, 0})
	assert.False(t, s.IsRebuyAvailable())
}

func TestFiveCardStud_IsAddonAvailable(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 3
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetAddonUsed([]bool{false, false})

	// Wrong hand number.
	s.SetHandCount(2)
	assert.False(t, s.IsAddonAvailable())

	// Correct hand number, addon unused.
	s.SetHandCount(3)
	assert.True(t, s.IsAddonAvailable())

	// Already used.
	s.SetAddonUsed([]bool{true, false})
	assert.False(t, s.IsAddonAvailable())

	// Addon disabled.
	s.SetAddonUsed([]bool{false, false})
	cfg.AddonEnabled = false
	s.SetConfig(cfg)
	assert.False(t, s.IsAddonAvailable())
}

func TestFiveCardStud_Reset_EntersRebuyPhase(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.RebuyEnabled = true
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	players[0].SetChips(0) // human broke -> must rebuy
	players[1].SetChips(0) // cpu broke -> auto-rebuy
	require.NoError(t, s.Reset())
	assert.Equal(t, FiveCardStudPhaseRebuy, s.GetPhase())
	assert.Equal(t, FiveCardStudRebuyPhaseRebuy, s.GetRebuyPhaseType())
	// CPU received automatic rebuy chips.
	assert.Greater(t, players[1].GetChips(), 0)
}

func TestFiveCardStud_Reset_TournamentAnteEscalation(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	cfg.TournamentMode = true
	cfg.AnteLevelHands = 1
	cfg.AnteMultiplier = 200 // doubles each level
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
		NewFiveCardStudPlayer(true, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	for _, p := range players {
		p.SetChips(10000)
	}
	s.SetHandCount(1) // handCount%AnteLevelHands == 0 triggers escalation
	require.NoError(t, s.Reset())
	// Ante should have doubled from 1 -> 2.
	assert.Equal(t, 2, s.GetConfig().Ante)
}

// --- Player getters / stats ---

func TestFiveCardStudPlayer_StatsAndGetters(t *testing.T) {
	p := NewFiveCardStudPlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.NotEmpty(t, p.GetPlayStyleName())

	// Cards.
	p.AddHoleCard(fcsCard(CardDesignSpade, 1))
	p.AddDoorCard(fcsCard(CardDesignHeart, 13))
	assert.Len(t, p.GetHoleCards(), 1)
	assert.Len(t, p.GetDoorCards(), 1)
	assert.Len(t, p.GetAllCards(), 2)
	p.ClearCards()
	assert.Empty(t, p.GetHoleCards())
	assert.Empty(t, p.GetDoorCards())

	// Hand stats.
	assert.Equal(t, 0, p.GetVPIP())
	assert.Equal(t, 0, p.GetPFR())
	assert.Equal(t, 0, p.GetThreeBet())
	p.IncrementTotalHands()
	p.IncrementTotalHands()
	p.IncrementVPIP()
	p.IncrementPFR()
	assert.Equal(t, 2, p.GetTotalHands())
	assert.Equal(t, 1, p.GetVPIPCount())
	assert.Equal(t, 1, p.GetPFRCount())
	assert.Equal(t, 50, p.GetVPIP())
	assert.Equal(t, 50, p.GetPFR())

	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBet()
	assert.Equal(t, 2, p.GetThreeBetOpportunity())
	assert.Equal(t, 1, p.GetThreeBetCount())
	assert.Equal(t, 50, p.GetThreeBet())
}

func TestFiveCardStudPlayer_AFDisplay(t *testing.T) {
	p := NewFiveCardStudPlayer(false, HoldemStyleGTO)
	assert.Equal(t, "-", p.GetAFDisplay()) // no action

	p.IncrementPostFlopBetRaise()
	p.IncrementPostFlopBetRaise()
	assert.Equal(t, "∞", p.GetAFDisplay()) // bet/raise but no call
	assert.Equal(t, 2, p.GetPostFlopBetRaise())

	p.IncrementPostFlopCall()
	assert.Equal(t, "2.0", p.GetAFDisplay())
	assert.Equal(t, 1, p.GetPostFlopCall())
}

func TestFiveCardStudPlayer_EvalBestHand_TooFewCards(t *testing.T) {
	p := NewFiveCardStudPlayer(false, HoldemStyleGTO)
	p.AddDoorCard(fcsCard(CardDesignSpade, 5))
	p.AddDoorCard(fcsCard(CardDesignHeart, 9))
	// < 5 cards -> high card, nil best hand.
	assert.Equal(t, PokerHandHighCard, p.EvalBestHand())
	assert.Nil(t, p.GetBestHand())
}

func TestFiveCardStudPlayer_EvalBestHand_Flush(t *testing.T) {
	p := NewFiveCardStudPlayer(false, HoldemStyleGTO)
	p.AddHoleCard(fcsCard(CardDesignSpade, 2))
	p.AddDoorCard(fcsCard(CardDesignSpade, 5))
	p.AddDoorCard(fcsCard(CardDesignSpade, 8))
	p.AddDoorCard(fcsCard(CardDesignSpade, 11))
	p.AddDoorCard(fcsCard(CardDesignSpade, 13))
	assert.Equal(t, PokerHandFlush, p.EvalBestHand())
	assert.Len(t, p.GetBestHand(), 5)
	assert.Len(t, p.GetComparisonCards(), 5)
}

func TestFiveCardStudPlayer_EvalVisibleHand(t *testing.T) {
	p := NewFiveCardStudPlayer(false, HoldemStyleGTO)
	// No door cards -> high card.
	assert.Equal(t, PokerHandHighCard, p.EvalVisibleHand())

	// A pair of door cards -> one pair.
	p.AddDoorCard(fcsCard(CardDesignSpade, 10))
	p.AddDoorCard(fcsCard(CardDesignHeart, 10))
	assert.Equal(t, PokerHandOnePair, p.EvalVisibleHand())

	// Five door cards -> full 5-card eval path.
	p.AddDoorCard(fcsCard(CardDesignClover, 10))
	p.AddDoorCard(fcsCard(CardDesignDiamond, 4))
	p.AddDoorCard(fcsCard(CardDesignSpade, 4))
	assert.Equal(t, PokerHandFullHouse, p.EvalVisibleHand())
}

func TestEvalPartialHandFcs(t *testing.T) {
	four := []*Card{
		fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 7),
		fcsCard(CardDesignClover, 7), fcsCard(CardDesignDiamond, 7),
	}
	assert.Equal(t, PokerHandFourOfAKind, evalPartialHandFcs(four))

	trips := []*Card{
		fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 7), fcsCard(CardDesignClover, 7),
	}
	assert.Equal(t, PokerHandThreeOfAKind, evalPartialHandFcs(trips))

	fullHouse := []*Card{
		fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 7), fcsCard(CardDesignClover, 7),
		fcsCard(CardDesignDiamond, 3), fcsCard(CardDesignSpade, 3),
	}
	assert.Equal(t, PokerHandFullHouse, evalPartialHandFcs(fullHouse))

	twoPair := []*Card{
		fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 7),
		fcsCard(CardDesignClover, 3), fcsCard(CardDesignDiamond, 3),
	}
	assert.Equal(t, PokerHandTwoPair, evalPartialHandFcs(twoPair))

	onePair := []*Card{fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 7)}
	assert.Equal(t, PokerHandOnePair, evalPartialHandFcs(onePair))

	high := []*Card{fcsCard(CardDesignSpade, 7), fcsCard(CardDesignHeart, 9)}
	assert.Equal(t, PokerHandHighCard, evalPartialHandFcs(high))

	assert.Equal(t, PokerHandHighCard, evalPartialHandFcs(nil))
}

func TestFiveCardStudPlayer_JSONRoundTrip(t *testing.T) {
	p := NewFiveCardStudPlayer(true, HoldemStyleLAG)
	p.SetChips(777)
	p.AddHoleCard(fcsCard(CardDesignSpade, 1))
	p.AddDoorCard(fcsCard(CardDesignHeart, 13))
	p.SetBestHand([]*Card{fcsCard(CardDesignSpade, 1)})
	p.SetTotalHands(5)
	p.SetVPIPCount(2)
	p.SetPFRCount(1)
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBet()
	p.IncrementPostFlopBetRaise()
	p.IncrementPostFlopCall()

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored FiveCardStudPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 777, restored.GetChips())
	assert.Equal(t, HoldemStyleLAG, restored.GetPlayStyle())
	assert.Len(t, restored.GetHoleCards(), 1)
	assert.Len(t, restored.GetDoorCards(), 1)
	assert.Equal(t, 5, restored.GetTotalHands())
	assert.Equal(t, 2, restored.GetVPIPCount())
	assert.Equal(t, 1, restored.GetThreeBetCount())
	assert.Equal(t, 1, restored.GetPostFlopBetRaise())
	assert.Equal(t, 1, restored.GetPostFlopCall())
}

func TestFiveCardStudPlayer_SettersForTest(t *testing.T) {
	p := NewFiveCardStudPlayer(false, HoldemStyleGTO)
	hole := []*Card{fcsCard(CardDesignSpade, 1)}
	door := []*Card{fcsCard(CardDesignHeart, 2), fcsCard(CardDesignClover, 3)}
	best := []*Card{fcsCard(CardDesignDiamond, 4)}
	p.SetHoleCards(hole)
	p.SetDoorCards(door)
	p.SetBestHand(best)
	p.SetPlayStyle(HoldemStyleTAP)
	p.SetTotalHands(3)
	p.SetVPIPCount(2)
	p.SetPFRCount(1)
	assert.Equal(t, hole, p.GetHoleCards())
	assert.Equal(t, door, p.GetDoorCards())
	assert.Equal(t, best, p.GetBestHand())
	assert.Equal(t, HoldemStyleTAP, p.GetPlayStyle())
	assert.Equal(t, 3, p.GetTotalHands())
	assert.Equal(t, 2, p.GetVPIPCount())
	assert.Equal(t, 1, p.GetPFRCount())
}

func TestFiveCardStud_SettersForTest(t *testing.T) {
	s := newTestFiveCardStud()
	s.SetSidePots([]SidePot{{Amount: 50}})
	assert.Equal(t, 50, s.GetSidePots()[0].Amount)
	s.SetCpuActions([]FiveCardStudCpuAction{{PlayerIdx: 1, Action: FiveCardStudActionCall}})
	assert.Len(t, s.GetCpuActions(), 1)
	s.SetCommunityCard(fcsCard(CardDesignSpade, 7))
	assert.NotNil(t, s.GetCommunityCard())
	s.SetLastHumanPlayMs(1234)
	assert.Equal(t, 1234, s.GetLastHumanPlayMs())
	s.SetRaiseCount(3)
	assert.Equal(t, 3, s.GetRaiseCount())
	s.SetHandCount(7)
	assert.Equal(t, 7, s.GetHandCount())
	prof := &BettingHumanProfile{}
	s.SetHumanProfile(prof)
	assert.Equal(t, prof, s.GetHumanProfile())
}

// --- Showdown / pots ---

func TestFiveCardStud_Showdown_FullHouseBeatsFlush(t *testing.T) {
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

	// Player 1: full house (sevens over threes).
	players[1].SetChips(1000)
	players[1].AddHoleCard(fcsCard(CardDesignSpade, 7))
	players[1].AddDoorCard(fcsCard(CardDesignHeart, 7))
	players[1].AddDoorCard(fcsCard(CardDesignClover, 7))
	players[1].AddDoorCard(fcsCard(CardDesignDiamond, 3))
	players[1].AddDoorCard(fcsCard(CardDesignSpade, 3))
	// Player 0: flush (all spades).
	players[0].SetChips(1000)
	players[0].AddHoleCard(fcsCard(CardDesignSpade, 2))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 5))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 8))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 11))
	players[0].AddDoorCard(fcsCard(CardDesignSpade, 13))

	s.SetActedFlags([]bool{true, true})
	s.resolveShowdown()

	results := s.GetRoundResults()
	require.NotEmpty(t, results)
	var p1Won int
	for _, r := range results {
		if r.PlayerIdx == 1 {
			p1Won = r.WonAmount
		}
	}
	assert.Greater(t, p1Won, 0) // full house wins the pot
	// Human (player 0) lost so phase stays at showdown (muck/show pending).
	assert.Equal(t, FiveCardStudPhaseShowdown, s.GetPhase())
	assert.True(t, s.IsMuckAvailable())
}

func TestFiveCardStud_Showdown_CommunityCardIncluded(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 2
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(false, HoldemStyleTAG),
		NewFiveCardStudPlayer(false, HoldemStyleTAG),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseFifthStreet)
	s.SetStartingChips([]int{1000, 1000})
	s.SetPot(100)
	// Each player has only 4 cards; a shared community card completes hands.
	for _, p := range players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, 9))
		p.AddDoorCard(fcsCard(CardDesignHeart, 8))
		p.AddDoorCard(fcsCard(CardDesignClover, 7))
		p.AddDoorCard(fcsCard(CardDesignDiamond, 6))
	}
	s.SetCommunityCard(fcsCard(CardDesignSpade, 5))
	s.SetActedFlags([]bool{true, true})
	s.resolveShowdown()

	// Community card was only temporarily added; hands trimmed back to 4 cards.
	assert.Len(t, players[0].GetHoleCards(), 1)
	assert.Len(t, players[0].GetDoorCards(), 3)
	// All-CPU showdown finalizes immediately.
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
	assert.True(t, s.GetGameEndFlag())
}

func TestFiveCardStud_ResolveLastPlayer(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 3
	players := []*FiveCardStudPlayer{
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
		NewFiveCardStudPlayer(false, HoldemStyleGTO),
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	for _, p := range players {
		p.SetChips(500)
	}
	s.SetPot(60)
	players[0].SetFolded(true)
	players[2].SetFolded(true)
	// Only player 1 remains.
	s.resolveLastPlayer()
	assert.Equal(t, 560, players[1].GetChips())
	assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
	assert.True(t, s.GetGameEndFlag())
	results := s.GetRoundResults()
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].PlayerIdx)
	assert.Equal(t, 60, results[0].WonAmount)
}

func TestFiveCardStud_GetHandName(t *testing.T) {
	s := newTestFiveCardStud()
	assert.Equal(t, "Unknown", s.getHandName(-1))
	assert.Equal(t, "Unknown", s.getHandName(9999))
	assert.Equal(t, PokerHandNames[PokerHandFlush], s.getHandName(PokerHandFlush))
}

// --- CPU decisions across streets/styles ---

func TestFiveCardStud_CpuDecide_AllStyles_AllStreets(t *testing.T) {
	styles := []FiveCardStudPlayStyle{
		HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO,
	}
	phases := []int{
		FiveCardStudPhaseSecondStreet, FiveCardStudPhaseThirdStreet,
		FiveCardStudPhaseFourthStreet, FiveCardStudPhaseFifthStreet,
	}
	for _, style := range styles {
		for _, phase := range phases {
			// Run several times so random branches (bluff/3bet) get exercised.
			for iter := 0; iter < 30; iter++ {
				s := newTestFiveCardStud()
				s.SetPhase(phase)
				s.SetPot(100)
				for _, p := range s.players {
					p.SetChips(1000)
					p.SetPlayStyle(style)
					p.AddHoleCard(fcsCard(CardDesignSpade, 13))
					p.AddDoorCard(fcsCard(CardDesignHeart, 13))
					if phase >= FiveCardStudPhaseThirdStreet {
						p.AddDoorCard(fcsCard(CardDesignClover, 9))
					}
					if phase >= FiveCardStudPhaseFourthStreet {
						p.AddDoorCard(fcsCard(CardDesignDiamond, 4))
					}
					if phase >= FiveCardStudPhaseFifthStreet {
						p.AddDoorCard(fcsCard(CardDesignSpade, 2))
					}
				}
				s.SetLastBet(5)
				s.SetRaiseCount(1)
				action, amount := s.cpuDecide(1)
				assert.GreaterOrEqual(t, action, FiveCardStudActionFold)
				assert.GreaterOrEqual(t, amount, 0)
			}
		}
	}
}

func TestFiveCardStud_CpuDecide_WeakHandNoBet(t *testing.T) {
	// Weak hand with no outstanding bet -> check/fold path.
	for _, style := range []FiveCardStudPlayStyle{HoldemStyleTAG, HoldemStyleTAP} {
		for iter := 0; iter < 30; iter++ {
			s := newTestFiveCardStud()
			s.SetPhase(FiveCardStudPhaseSecondStreet)
			for _, p := range s.players {
				p.SetChips(1000)
				p.SetPlayStyle(style)
				p.AddHoleCard(fcsCard(CardDesignSpade, 2))
				p.AddDoorCard(fcsCard(CardDesignHeart, 7))
			}
			s.SetLastBet(0)
			action, _ := s.cpuDecide(1)
			assert.GreaterOrEqual(t, action, FiveCardStudActionFold)
		}
	}
}

func TestFiveCardStud_CpuDecide_PotLimitClampsBet(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	cfg.BettingLimit = BettingLimitPotLimit
	players := make([]*FiveCardStudPlayer, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, NewFiveCardStudPlayer(false, HoldemStyleLAG))
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetPhase(FiveCardStudPhaseSecondStreet)
	s.SetPot(20)
	for _, p := range players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, 1))
		p.AddDoorCard(fcsCard(CardDesignSpade, 1))
	}
	s.SetLastBet(5)
	s.SetRaiseCount(4) // at the fixed raise cap -> raise downgraded to call
	action, amount := s.cpuDecide(1)
	assert.GreaterOrEqual(t, action, FiveCardStudActionFold)
	assert.GreaterOrEqual(t, amount, 0)
}

func TestFiveCardStud_CpuDecide_MetaAI(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	cfg.CpuMetaAI = true
	players := make([]*FiveCardStudPlayer, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, NewFiveCardStudPlayer(false, HoldemStyleTAP))
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	s.SetHumanProfile(&BettingHumanProfile{})
	s.SetLastHumanPlayMs(2000)
	s.SetPhase(FiveCardStudPhaseThirdStreet)
	s.SetPot(50)
	for _, p := range players {
		p.SetChips(1000)
		p.AddHoleCard(fcsCard(CardDesignSpade, 2))
		p.AddDoorCard(fcsCard(CardDesignHeart, 7))
		p.AddDoorCard(fcsCard(CardDesignClover, 4))
	}
	s.SetLastBet(10)
	for iter := 0; iter < 30; iter++ {
		action, _ := s.cpuDecide(1)
		assert.GreaterOrEqual(t, action, FiveCardStudActionFold)
	}
}

func TestFiveCardStud_EvalThirdStreetStrength_Variants(t *testing.T) {
	s := newTestFiveCardStud()

	// Trips (very strong).
	s.players[0].AddHoleCard(fcsCard(CardDesignSpade, 8))
	s.players[0].AddDoorCard(fcsCard(CardDesignHeart, 8))
	s.players[0].AddDoorCard(fcsCard(CardDesignClover, 8))
	assert.Greater(t, s.evalThirdStreetStrength(0), 80)

	// Suited connectors (no pair) get a bonus.
	s.players[1].AddHoleCard(fcsCard(CardDesignDiamond, 9))
	s.players[1].AddDoorCard(fcsCard(CardDesignDiamond, 10))
	s.players[1].AddDoorCard(fcsCard(CardDesignDiamond, 11))
	connStrength := s.evalThirdStreetStrength(1)
	assert.Greater(t, connStrength, 0)

	// Single card -> 0.
	s.players[2].AddHoleCard(fcsCard(CardDesignSpade, 5))
	assert.Equal(t, 0, s.evalThirdStreetStrength(2))
}

// --- full all-CPU game variations to drive FiveCardStud + CPU branches ---

func TestFiveCardStud_FullCPUGame_PotLimit(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 4
	cfg.BettingLimit = BettingLimitPotLimit
	players := make([]*FiveCardStudPlayer, 0, 4)
	styles := DefaultFiveCardStudCpuStyles(4)
	for i := 0; i < 4; i++ {
		players = append(players, NewFiveCardStudPlayer(false, styles[i%len(styles)]))
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	for hand := 0; hand < 5; hand++ {
		require.NoError(t, s.Reset())
		assert.True(t, s.GetGameEndFlag())
		assert.Equal(t, FiveCardStudPhaseEnd, s.GetPhase())
		require.NoError(t, s.GetLastCpuError())
	}
}

func TestFiveCardStud_FullCPUGame_Tournament(t *testing.T) {
	cfg := DefaultFiveCardStudConfig()
	cfg.TableSize = 3
	cfg.TournamentMode = true
	cfg.AnteLevelHands = 2
	cfg.AnteMultiplier = 150
	players := make([]*FiveCardStudPlayer, 0, 3)
	styles := DefaultFiveCardStudCpuStyles(3)
	players = append(players, NewFiveCardStudPlayer(false, HoldemStyleGTO))
	for _, st := range styles {
		players = append(players, NewFiveCardStudPlayer(false, st))
	}
	s := NewFiveCardStud(NewTrumpCards(0), players, cfg)
	for hand := 0; hand < 6; hand++ {
		require.NoError(t, s.Reset())
		require.NoError(t, s.GetLastCpuError())
	}
}
