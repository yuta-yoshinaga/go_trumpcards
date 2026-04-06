//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSevenCardStud() *SevenCardStud {
	cfg := DefaultSevenCardStudConfig()
	cfg.TableSize = 4
	players := NewSevenCardStudPlayersForTable(cfg.TableSize)
	tc := NewTrumpCards(0)
	return NewSevenCardStud(tc, players, cfg)
}

// setupSevenCardStudForHumanAction creates a game at the given phase with human at currentTurn.
func setupSevenCardStudForHumanAction(phase int) *SevenCardStud {
	s := newTestSevenCardStud()
	s.SetPhase(phase)
	s.SetCurrentTurn(0) // human is player 0
	s.SetPot(20)
	s.SetLastBet(5)
	s.SetMinRaise(5)

	// Give each player chips and cards
	for i, p := range s.players {
		p.SetChips(1000)
		p.AddHoleCard(NewCard(CardDesignSpade, i+1, true))
		p.AddHoleCard(NewCard(CardDesignHeart, i+1, true))
		p.AddDoorCard(NewCard(CardDesignDiamond, i+5, true))
		s.SetActedFlags(make([]bool, len(s.players)))
		s.SetStartingChips([]int{1000, 1000, 1000, 1000})
	}
	return s
}

func TestNewSevenCardStud(t *testing.T) {
	s := newTestSevenCardStud()
	assert.Equal(t, SevenCardStudPhaseInit, s.GetPhase())
	assert.Equal(t, 4, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetPot())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, -1, s.GetBringInPlayerIdx())
}

func TestSevenCardStud_Reset(t *testing.T) {
	s := newTestSevenCardStud()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	err := s.Reset()
	require.NoError(t, err)

	// After reset: phase should be ThirdStreet (or End if all CPU folded)
	assert.True(t, s.GetPhase() >= SevenCardStudPhaseThirdStreet)

	// All players should have 3 cards (2 hole + 1 door) minus ante
	for _, p := range s.players {
		assert.Len(t, p.GetHoleCards(), 2)
		assert.Len(t, p.GetDoorCards(), 1)
	}

	// Pot should have antes + bring-in
	assert.Greater(t, s.GetPot(), 0)
	assert.GreaterOrEqual(t, s.GetBringInPlayerIdx(), 0)
}

func TestSevenCardStud_PlayerAction_Fold(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	err := s.PlayerAction(SevenCardStudActionFold, 0, 0)
	require.NoError(t, err)
}

func TestSevenCardStud_PlayerAction_Call(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	err := s.PlayerAction(SevenCardStudActionCall, 0, 100)
	require.NoError(t, err)
}

func TestSevenCardStud_PlayerAction_Check(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	s.SetLastBet(0) // no bet to call
	err := s.PlayerAction(SevenCardStudActionCheck, 0, 0)
	require.NoError(t, err)
}

func TestSevenCardStud_PlayerAction_Raise(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	err := s.PlayerAction(SevenCardStudActionRaise, 10, 0)
	require.NoError(t, err)
}

func TestSevenCardStud_PlayerAction_Errors(t *testing.T) {
	t.Run("game ended", func(t *testing.T) {
		s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
		s.SetGameEndFlag(true)
		err := s.PlayerAction(SevenCardStudActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("wrong phase - init", func(t *testing.T) {
		s := setupSevenCardStudForHumanAction(SevenCardStudPhaseInit)
		err := s.PlayerAction(SevenCardStudActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("wrong phase - showdown", func(t *testing.T) {
		s := setupSevenCardStudForHumanAction(SevenCardStudPhaseShowdown)
		err := s.PlayerAction(SevenCardStudActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("not human turn", func(t *testing.T) {
		s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
		s.SetCurrentTurn(1) // CPU player
		err := s.PlayerAction(SevenCardStudActionFold, 0, 0)
		assert.Error(t, err)
	})
}

func TestSevenCardStud_DetermineBringIn(t *testing.T) {
	s := newTestSevenCardStud()

	// Give each player a door card
	s.players[0].AddDoorCard(NewCard(CardDesignSpade, 10, true))  // 10♠
	s.players[1].AddDoorCard(NewCard(CardDesignHeart, 3, true))   // 3♥ (lowest)
	s.players[2].AddDoorCard(NewCard(CardDesignDiamond, 7, true)) // 7♦
	s.players[3].AddDoorCard(NewCard(CardDesignClover, 13, true)) // K♣

	idx := s.determineBringIn()
	assert.Equal(t, 1, idx) // 3♥ is the lowest
}

func TestSevenCardStud_DetermineBringIn_TieBreakBySuit(t *testing.T) {
	s := newTestSevenCardStud()

	// Two players with same value door cards
	s.players[0].AddDoorCard(NewCard(CardDesignSpade, 5, true))    // 5♠ (suit rank 4)
	s.players[1].AddDoorCard(NewCard(CardDesignClover, 5, true))   // 5♣ (suit rank 1, lowest)
	s.players[2].AddDoorCard(NewCard(CardDesignHeart, 5, true))    // 5♥ (suit rank 3)
	s.players[3].AddDoorCard(NewCard(CardDesignDiamond, 10, true)) // 10♦

	idx := s.determineBringIn()
	assert.Equal(t, 1, idx) // 5♣ has lowest suit rank
}

func TestSevenCardStud_DetermineBettingLeader(t *testing.T) {
	s := newTestSevenCardStud()
	s.SetPhase(SevenCardStudPhaseFourthStreet)

	// Player 0: pair of 10s showing
	s.players[0].AddDoorCard(NewCard(CardDesignSpade, 10, true))
	s.players[0].AddDoorCard(NewCard(CardDesignHeart, 10, true))

	// Player 1: high card
	s.players[1].AddDoorCard(NewCard(CardDesignDiamond, 1, true))
	s.players[1].AddDoorCard(NewCard(CardDesignClover, 5, true))

	// Player 2: high card
	s.players[2].AddDoorCard(NewCard(CardDesignSpade, 8, true))
	s.players[2].AddDoorCard(NewCard(CardDesignHeart, 7, true))

	// Player 3 folded
	s.players[3].SetFolded(true)

	leader := s.determineBettingLeader()
	assert.Equal(t, 0, leader) // pair of 10s is best visible
}

func TestSevenCardStud_Muck(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseShowdown)
	s.SetRoundResults([]SevenCardStudResult{
		{PlayerIdx: 0, WonAmount: 0},
		{PlayerIdx: 1, WonAmount: 50},
	})

	assert.True(t, s.IsMuckAvailable())

	err := s.Muck()
	require.NoError(t, err)
	assert.Equal(t, SevenCardStudPhaseEnd, s.GetPhase())
	assert.True(t, s.roundResults[0].Mucked)
}

func TestSevenCardStud_ShowHand(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseShowdown)
	err := s.ShowHand()
	require.NoError(t, err)
	assert.Equal(t, SevenCardStudPhaseEnd, s.GetPhase())
}

func TestSevenCardStud_Muck_WrongPhase(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	err := s.Muck()
	assert.Error(t, err)
}

func TestSevenCardStud_IsHumanTurn(t *testing.T) {
	s := newTestSevenCardStud()
	s.SetCurrentTurn(0)
	assert.True(t, s.IsHumanTurn())
	s.SetCurrentTurn(1)
	assert.False(t, s.IsHumanTurn())
}

func TestSevenCardStud_Getters(t *testing.T) {
	s := newTestSevenCardStud()
	s.SetPhase(SevenCardStudPhaseThirdStreet)
	s.SetPot(100)
	s.SetDealerIdx(2)
	s.SetCurrentTurn(1)
	s.SetLastBet(10)
	s.SetMinRaise(5)
	s.SetGameEndFlag(false)
	s.SetBringInPlayerIdx(3)

	assert.Equal(t, SevenCardStudPhaseThirdStreet, s.GetPhase())
	assert.Equal(t, 100, s.GetPot())
	assert.Equal(t, 2, s.GetDealerIdx())
	assert.Equal(t, 1, s.GetCurrentTurn())
	assert.Equal(t, 10, s.GetLastBet())
	assert.Equal(t, 5, s.GetMinRaise())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, 3, s.GetBringInPlayerIdx())

	cfg := s.GetConfig()
	assert.Equal(t, 4, cfg.TableSize)

	assert.NotNil(t, s.GetPlayer(0))
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(99))
}

func TestSevenCardStud_Config(t *testing.T) {
	s := newTestSevenCardStud()
	newCfg := DefaultSevenCardStudConfig()
	newCfg.Ante = 5
	s.SetConfig(newCfg)
	assert.Equal(t, 5, s.GetConfig().Ante)
}

func TestSevenCardStud_Resize(t *testing.T) {
	s := newTestSevenCardStud()
	newPlayers := NewSevenCardStudPlayersForTable(3)
	s.Resize(newPlayers)
	assert.Equal(t, 3, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetHandCount())
}

func TestSevenCardStud_JSON(t *testing.T) {
	s := newTestSevenCardStud()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	err := s.Reset()
	require.NoError(t, err)

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var restored SevenCardStud
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetPot(), restored.GetPot())
	assert.Equal(t, s.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, s.GetHandCount(), restored.GetHandCount())
	assert.Equal(t, s.GetBringInPlayerIdx(), restored.GetBringInPlayerIdx())
}

func TestSevenCardStud_JSON_MaxSlice(t *testing.T) {
	// Test that oversized arrays are rejected
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
	var restored SevenCardStud
	err := json.Unmarshal([]byte(badJSON), &restored)
	assert.Error(t, err)
}

func TestSevenCardStud_FullGame(t *testing.T) {
	// Run multiple resets to exercise the full game loop
	s := newTestSevenCardStud()
	for _, p := range s.players {
		p.SetChips(1000)
	}

	for i := 0; i < 5; i++ {
		err := s.Reset()
		require.NoError(t, err)

		// If human's turn, fold to move the game forward
		if s.IsHumanTurn() && s.GetPhase() >= SevenCardStudPhaseThirdStreet && s.GetPhase() <= SevenCardStudPhaseSeventhStreet {
			err = s.PlayerAction(SevenCardStudActionFold, 0, 0)
			require.NoError(t, err)
		}
	}
}

func TestSevenCardStud_CurrentBetSize(t *testing.T) {
	s := newTestSevenCardStud()

	s.SetPhase(SevenCardStudPhaseThirdStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(SevenCardStudPhaseFourthStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(SevenCardStudPhaseFifthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())

	s.SetPhase(SevenCardStudPhaseSixthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())

	s.SetPhase(SevenCardStudPhaseSeventhStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())
}

func TestSevenCardStud_Rebuy(t *testing.T) {
	t.Run("rebuy wrong phase", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.SetPhase(SevenCardStudPhaseThirdStreet)
		assert.Error(t, s.Rebuy())
	})

	t.Run("skip rebuy wrong phase", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.SetPhase(SevenCardStudPhaseThirdStreet)
		assert.Error(t, s.SkipRebuy())
	})

	t.Run("addon wrong phase", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.SetPhase(SevenCardStudPhaseThirdStreet)
		assert.Error(t, s.Addon())
	})

	t.Run("skip addon wrong phase", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.SetPhase(SevenCardStudPhaseThirdStreet)
		assert.Error(t, s.SkipAddon())
	})

	t.Run("rebuy happy path", func(t *testing.T) {
		s := newTestSevenCardStud()
		for _, p := range s.players {
			p.SetChips(0)
		}
		cfg := s.GetConfig()
		cfg.RebuyEnabled = true
		cfg.RebuyMaxCount = 3
		cfg.RebuyChips = 500
		cfg.RebuyPeriodHands = 20
		s.SetConfig(cfg)
		s.SetPhase(SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(SevenCardStudRebuyPhaseRebuy)
		s.SetRebuyCounts(make([]int, len(s.GetPlayers())))

		err := s.Rebuy()
		require.NoError(t, err)
		// After rebuy (500 chips), continueReset() deducts ante (1 chip)
		assert.Equal(t, 499, s.GetPlayers()[0].GetChips())
	})

	t.Run("skip rebuy - human busted", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.GetPlayers()[0].SetChips(0)
		s.SetPhase(SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(SevenCardStudRebuyPhaseRebuy)

		err := s.SkipRebuy()
		require.NoError(t, err)
		assert.Equal(t, SevenCardStudPhaseEnd, s.GetPhase())
		assert.True(t, s.GetGameEndFlag())
	})

	t.Run("addon happy path", func(t *testing.T) {
		s := newTestSevenCardStud()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		cfg := s.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonChips = 500
		cfg.AddonAfterHand = 1
		s.SetConfig(cfg)
		s.SetPhase(SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(SevenCardStudRebuyPhaseAddon)
		s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

		err := s.Addon()
		require.NoError(t, err)
		// After addon, continueReset() runs which deducts ante (1 chip)
		assert.Equal(t, 1499, s.GetPlayers()[0].GetChips())
	})

	t.Run("skip addon happy path", func(t *testing.T) {
		s := newTestSevenCardStud()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		s.SetPhase(SevenCardStudPhaseRebuy)
		s.SetRebuyPhaseType(SevenCardStudRebuyPhaseAddon)
		s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

		err := s.SkipAddon()
		require.NoError(t, err)
		// After skip addon, continueReset() runs which deducts ante
		assert.Less(t, s.GetPlayers()[0].GetChips(), 1000)
	})

	t.Run("IsRebuyAvailable", func(t *testing.T) {
		s := newTestSevenCardStud()
		assert.False(t, s.IsRebuyAvailable())

		cfg := s.GetConfig()
		cfg.RebuyEnabled = true
		cfg.RebuyMaxCount = 3
		cfg.RebuyPeriodHands = 20
		s.SetConfig(cfg)
		s.SetHandCount(1)
		s.SetRebuyCounts(make([]int, len(s.GetPlayers())))
		s.GetPlayers()[0].SetChips(0)
		assert.True(t, s.IsRebuyAvailable())
	})

	t.Run("IsAddonAvailable", func(t *testing.T) {
		s := newTestSevenCardStud()
		assert.False(t, s.IsAddonAvailable())

		cfg := s.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonAfterHand = 5
		s.SetConfig(cfg)
		s.SetHandCount(5)
		s.SetAddonUsed(make([]bool, len(s.GetPlayers())))
		assert.True(t, s.IsAddonAvailable())
	})

	t.Run("GetRebuyCounts and GetAddonUsed", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.SetRebuyCounts([]int{1, 0, 0, 0})
		s.SetAddonUsed([]bool{true, false, false, false})
		assert.Equal(t, []int{1, 0, 0, 0}, s.GetRebuyCounts())
		assert.Equal(t, []bool{true, false, false, false}, s.GetAddonUsed())
		assert.Equal(t, SevenCardStudRebuyPhaseNone, s.GetRebuyPhaseType())
	})

	t.Run("rebuy with tournament reset", func(t *testing.T) {
		s := newTestSevenCardStud()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		cfg := s.GetConfig()
		cfg.RebuyEnabled = true
		cfg.RebuyMaxCount = 1
		cfg.RebuyChips = 1000
		cfg.RebuyPeriodHands = 50
		s.SetConfig(cfg)

		// Simulate: human has chips, play a hand, then drain to 0
		err := s.Reset()
		require.NoError(t, err)
		// Human folds
		if s.IsHumanTurn() {
			_ = s.PlayerAction(SevenCardStudActionFold, 0, 0)
		}
		// Drain human chips to trigger rebuy on next reset
		s.GetPlayers()[0].SetChips(0)
		err = s.Reset()
		require.NoError(t, err)
		// Should be in rebuy phase or have continued
		if s.GetPhase() == SevenCardStudPhaseRebuy {
			assert.Equal(t, SevenCardStudRebuyPhaseRebuy, s.GetRebuyPhaseType())
		}
	})
}

func TestSevenCardStud_CommunityCardShowdown(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseShowdown)
	// Give each non-folded player 7 cards for proper evaluation
	for _, p := range s.players {
		for len(p.GetHoleCards()) < 3 {
			p.AddHoleCard(NewCard(CardDesignClover, 2, true))
		}
		for len(p.GetDoorCards()) < 4 {
			p.AddDoorCard(NewCard(CardDesignDiamond, 3, true))
		}
	}
	s.SetStartingChips([]int{1000, 1000, 1000, 1000})

	// Set a community card
	cc := NewCard(CardDesignSpade, 1, true)
	s.SetCommunityCard(cc)

	s.resolveShowdown()

	// Verify community card was NOT permanently added to any player's holeCards
	for i, p := range s.players {
		assert.Len(t, p.GetHoleCards(), 3, "player %d should still have 3 hole cards", i)
	}
}

func TestSevenCardStud_CPUDecide_AllStyles(t *testing.T) {
	styles := []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	for _, style := range styles {
		t.Run(HoldemPlayStyleNames[style], func(t *testing.T) {
			s := newTestSevenCardStud()
			for _, p := range s.players {
				p.SetChips(1000)
			}
			// Set CPU player style
			s.GetPlayers()[1].SetPlayStyle(style)

			// Give player 1 cards for evaluation
			s.GetPlayers()[1].AddHoleCard(NewCard(CardDesignSpade, 10, true))
			s.GetPlayers()[1].AddHoleCard(NewCard(CardDesignHeart, 10, true))
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignDiamond, 5, true))

			s.SetPhase(SevenCardStudPhaseThirdStreet)
			s.SetLastBet(5)
			s.SetMinRaise(5)

			action, amount := s.cpuDecide(1)
			assert.GreaterOrEqual(t, action, 0)
			assert.GreaterOrEqual(t, amount, 0)

			// Also test post-third-street
			s.SetPhase(SevenCardStudPhaseFifthStreet)
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignClover, 8, true))
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignSpade, 9, true))

			action2, amount2 := s.cpuDecide(1)
			assert.GreaterOrEqual(t, action2, 0)
			assert.GreaterOrEqual(t, amount2, 0)
		})
	}
}

func TestSevenCardStud_CPUDecide_MetaAI(t *testing.T) {
	s := newTestSevenCardStud()
	cfg := s.GetConfig()
	cfg.CpuMetaAI = true
	s.SetConfig(cfg)
	s.SetHumanProfile(&BettingHumanProfile{GamesPlayed: 10})

	for _, p := range s.players {
		p.SetChips(1000)
		p.AddHoleCard(NewCard(CardDesignSpade, 10, true))
		p.AddHoleCard(NewCard(CardDesignHeart, 5, true))
		p.AddDoorCard(NewCard(CardDesignDiamond, 7, true))
	}

	s.SetPhase(SevenCardStudPhaseThirdStreet)
	s.SetLastBet(5)
	s.SetMinRaise(5)

	// Should not panic with meta AI enabled
	action, amount := s.cpuDecide(1)
	assert.GreaterOrEqual(t, action, 0)
	assert.GreaterOrEqual(t, amount, 0)
}

func TestSevenCardStud_HandleCpuActionError_NoCallAmount(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	s.SetCurrentTurn(1)
	s.SetLastBet(0) // no call amount → should check instead of fold

	s.handleCpuActionError(1, SevenCardStudActionBet, assert.AnError)
	assert.NotNil(t, s.GetLastCpuError())
}

func TestSevenCardStud_EvalThirdStreetStrength_Variants(t *testing.T) {
	t.Run("suited cards", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 10, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 11, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignSpade, 12, true))

		strength := s.evalThirdStreetStrength(0)
		assert.GreaterOrEqual(t, strength, 40) // suited connected high cards
	})

	t.Run("low disconnected", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 2, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignHeart, 7, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignDiamond, 12, true))

		strength := s.evalThirdStreetStrength(0)
		assert.Greater(t, strength, 0)
	})

	t.Run("pair", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 8, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignHeart, 8, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignDiamond, 3, true))

		strength := s.evalThirdStreetStrength(0)
		assert.GreaterOrEqual(t, strength, 40)
	})

	t.Run("fewer than 3 cards", func(t *testing.T) {
		s := newTestSevenCardStud()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 8, true))

		strength := s.evalThirdStreetStrength(0)
		assert.Equal(t, 0, strength)
	})
}

func TestSevenCardStud_BettingLimits(t *testing.T) {
	s := newTestSevenCardStud()
	s.SetPot(100)
	s.SetLastBet(10)

	maxRaises, maxBetAmount := s.bettingLimits()
	assert.Greater(t, maxRaises, 0) // Fixed limit
	assert.Equal(t, 0, maxBetAmount)

	cfg := s.GetConfig()
	cfg.BettingLimit = BettingLimitNoLimit
	s.SetConfig(cfg)
	maxRaises, maxBetAmount = s.bettingLimits()
	assert.Equal(t, 0, maxRaises)
	assert.Equal(t, 0, maxBetAmount)
}

func TestSevenCardStud_MetaAI(t *testing.T) {
	s := newTestSevenCardStud()
	assert.Nil(t, s.GetHumanProfile())

	s.ResetProfile()
	assert.Nil(t, s.GetHumanProfile())

	assert.Nil(t, s.ExportProfile())
	assert.NoError(t, s.ImportProfile(nil))
	assert.NoError(t, s.ImportProfile([]byte{}))
}

func TestSevenCardStud_ActionLog(t *testing.T) {
	s := newTestSevenCardStud()
	assert.Empty(t, s.GetActionLog())

	s.appendLog(0, "test", "test action", nil)
	assert.Len(t, s.GetActionLog(), 1)
	assert.Equal(t, 1, s.GetActionLog()[0].TurnNumber)
}

func TestSevenCardStud_CountActivePlayers(t *testing.T) {
	s := newTestSevenCardStud()
	assert.Equal(t, 4, s.countActivePlayers())

	s.players[1].SetFolded(true)
	assert.Equal(t, 3, s.countActivePlayers())

	s.players[2].SetFolded(true)
	s.players[3].SetFolded(true)
	assert.Equal(t, 1, s.countActivePlayers())
}

func TestSevenCardStud_EvalThirdStreetStrength(t *testing.T) {
	s := newTestSevenCardStud()

	// Trip aces
	s.players[0].AddHoleCard(NewCard(CardDesignSpade, 1, true))
	s.players[0].AddHoleCard(NewCard(CardDesignHeart, 1, true))
	s.players[0].AddDoorCard(NewCard(CardDesignDiamond, 1, true))

	strength := s.evalThirdStreetStrength(0)
	assert.GreaterOrEqual(t, strength, 80) // trips should be very strong
}

func TestSevenCardStud_HandleCpuActionError(t *testing.T) {
	s := setupSevenCardStudForHumanAction(SevenCardStudPhaseThirdStreet)
	s.SetCurrentTurn(1) // CPU player

	// With call amount > 0, should fold
	s.handleCpuActionError(1, SevenCardStudActionBet, assert.AnError)
	assert.NotNil(t, s.GetLastCpuError())
}

func TestSevenCardStud_TournamentEscalation(t *testing.T) {
	s := newTestSevenCardStud()
	cfg := s.GetConfig()
	cfg.TournamentMode = true
	cfg.AnteLevelHands = 1
	cfg.AnteMultiplier = 200
	s.SetConfig(cfg)

	for _, p := range s.players {
		p.SetChips(10000)
	}

	// First reset
	err := s.Reset()
	require.NoError(t, err)
	initialAnte := s.GetConfig().Ante

	// Fold human
	if s.IsHumanTurn() {
		_ = s.PlayerAction(SevenCardStudActionFold, 0, 0)
	}

	// Second reset should escalate
	err = s.Reset()
	require.NoError(t, err)
	assert.Greater(t, s.GetConfig().Ante, initialAnte)
}
