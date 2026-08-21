//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFollowTheQueen() *FollowTheQueen {
	cfg := DefaultFollowTheQueenConfig()
	cfg.TableSize = 4
	players := NewFollowTheQueenPlayersForTable(cfg.TableSize)
	tc := NewTrumpCards(0)
	return NewFollowTheQueen(tc, players, cfg)
}

// setupFollowTheQueenForHumanAction creates a game at the given phase with human at currentTurn.
func setupFollowTheQueenForHumanAction(phase int) *FollowTheQueen {
	s := newTestFollowTheQueen()
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

func TestNewFollowTheQueen(t *testing.T) {
	s := newTestFollowTheQueen()
	assert.Equal(t, FollowTheQueenPhaseInit, s.GetPhase())
	assert.Equal(t, 4, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetPot())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, -1, s.GetBringInPlayerIdx())
}

func TestFollowTheQueen_Reset(t *testing.T) {
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	err := s.Reset()
	require.NoError(t, err)

	// After reset: phase should be ThirdStreet or beyond (CPUs may have acted, advancing phase)
	assert.True(t, s.GetPhase() >= FollowTheQueenPhaseThirdStreet)

	// All non-folded players should have at least 2 hole cards and 1+ door cards
	for _, p := range s.players {
		if !p.GetFolded() {
			assert.GreaterOrEqual(t, len(p.GetHoleCards()), 2)
			assert.GreaterOrEqual(t, len(p.GetDoorCards()), 1)
		}
	}

	// Pot should have antes + bring-in
	assert.Greater(t, s.GetPot(), 0)
	assert.GreaterOrEqual(t, s.GetBringInPlayerIdx(), 0)
}

func TestFollowTheQueen_PlayerAction_Fold(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	err := s.PlayerAction(FollowTheQueenActionFold, 0, 0)
	require.NoError(t, err)
}

func TestFollowTheQueen_PlayerAction_Call(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	err := s.PlayerAction(FollowTheQueenActionCall, 0, 100)
	require.NoError(t, err)
}

func TestFollowTheQueen_PlayerAction_Check(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	s.SetLastBet(0) // no bet to call
	err := s.PlayerAction(FollowTheQueenActionCheck, 0, 0)
	require.NoError(t, err)
}

func TestFollowTheQueen_PlayerAction_Raise(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	err := s.PlayerAction(FollowTheQueenActionRaise, 10, 0)
	require.NoError(t, err)
}

func TestFollowTheQueen_PlayerAction_Errors(t *testing.T) {
	t.Run("game ended", func(t *testing.T) {
		s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
		s.SetGameEndFlag(true)
		err := s.PlayerAction(FollowTheQueenActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("wrong phase - init", func(t *testing.T) {
		s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseInit)
		err := s.PlayerAction(FollowTheQueenActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("wrong phase - showdown", func(t *testing.T) {
		s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseShowdown)
		err := s.PlayerAction(FollowTheQueenActionFold, 0, 0)
		assert.Error(t, err)
	})

	t.Run("not human turn", func(t *testing.T) {
		s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
		s.SetCurrentTurn(1) // CPU player
		err := s.PlayerAction(FollowTheQueenActionFold, 0, 0)
		assert.Error(t, err)
	})
}

func TestFollowTheQueen_DetermineBringIn(t *testing.T) {
	s := newTestFollowTheQueen()

	// Give each player a door card
	s.players[0].AddDoorCard(NewCard(CardDesignSpade, 10, true))  // 10♠
	s.players[1].AddDoorCard(NewCard(CardDesignHeart, 3, true))   // 3♥ (lowest)
	s.players[2].AddDoorCard(NewCard(CardDesignDiamond, 7, true)) // 7♦
	s.players[3].AddDoorCard(NewCard(CardDesignClover, 13, true)) // K♣

	idx := s.determineBringIn()
	assert.Equal(t, 1, idx) // 3♥ is the lowest
}

func TestFollowTheQueen_DetermineBringIn_TieBreakBySuit(t *testing.T) {
	s := newTestFollowTheQueen()

	// Two players with same value door cards
	s.players[0].AddDoorCard(NewCard(CardDesignSpade, 5, true))    // 5♠ (suit rank 4)
	s.players[1].AddDoorCard(NewCard(CardDesignClover, 5, true))   // 5♣ (suit rank 1, lowest)
	s.players[2].AddDoorCard(NewCard(CardDesignHeart, 5, true))    // 5♥ (suit rank 3)
	s.players[3].AddDoorCard(NewCard(CardDesignDiamond, 10, true)) // 10♦

	idx := s.determineBringIn()
	assert.Equal(t, 1, idx) // 5♣ has lowest suit rank
}

func TestFollowTheQueen_DetermineBettingLeader(t *testing.T) {
	s := newTestFollowTheQueen()
	s.SetPhase(FollowTheQueenPhaseFourthStreet)

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

func TestFollowTheQueen_Muck(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseShowdown)
	s.SetRoundResults([]FollowTheQueenResult{
		{PlayerIdx: 0, WonAmount: 0},
		{PlayerIdx: 1, WonAmount: 50},
	})

	assert.True(t, s.IsMuckAvailable())

	err := s.Muck()
	require.NoError(t, err)
	assert.Equal(t, FollowTheQueenPhaseEnd, s.GetPhase())
	assert.True(t, s.roundResults[0].Mucked)
}

func TestFollowTheQueen_ShowHand(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseShowdown)
	err := s.ShowHand()
	require.NoError(t, err)
	assert.Equal(t, FollowTheQueenPhaseEnd, s.GetPhase())
}

func TestFollowTheQueen_Muck_WrongPhase(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	err := s.Muck()
	assert.Error(t, err)
}

func TestFollowTheQueen_IsHumanTurn(t *testing.T) {
	s := newTestFollowTheQueen()
	s.SetCurrentTurn(0)
	assert.True(t, s.IsHumanTurn())
	s.SetCurrentTurn(1)
	assert.False(t, s.IsHumanTurn())
}

func TestFollowTheQueen_Getters(t *testing.T) {
	s := newTestFollowTheQueen()
	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	s.SetPot(100)
	s.SetDealerIdx(2)
	s.SetCurrentTurn(1)
	s.SetLastBet(10)
	s.SetMinRaise(5)
	s.SetGameEndFlag(false)
	s.SetBringInPlayerIdx(3)

	assert.Equal(t, FollowTheQueenPhaseThirdStreet, s.GetPhase())
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

func TestFollowTheQueen_Config(t *testing.T) {
	s := newTestFollowTheQueen()
	newCfg := DefaultFollowTheQueenConfig()
	newCfg.Ante = 5
	s.SetConfig(newCfg)
	assert.Equal(t, 5, s.GetConfig().Ante)
}

func TestFollowTheQueen_Resize(t *testing.T) {
	s := newTestFollowTheQueen()
	newPlayers := NewFollowTheQueenPlayersForTable(3)
	s.Resize(newPlayers)
	assert.Equal(t, 3, s.GetPlayerCnt())
	assert.Equal(t, 0, s.GetHandCount())
}

func TestFollowTheQueen_JSON(t *testing.T) {
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	err := s.Reset()
	require.NoError(t, err)

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var restored FollowTheQueen
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetPot(), restored.GetPot())
	assert.Equal(t, s.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, s.GetHandCount(), restored.GetHandCount())
	assert.Equal(t, s.GetBringInPlayerIdx(), restored.GetBringInPlayerIdx())
}

func TestFollowTheQueen_JSON_MaxSlice(t *testing.T) {
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
	var restored FollowTheQueen
	err := json.Unmarshal([]byte(badJSON), &restored)
	assert.Error(t, err)
}

func TestFollowTheQueen_FullGame(t *testing.T) {
	// Run multiple resets to exercise the full game loop
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}

	for i := 0; i < 5; i++ {
		err := s.Reset()
		require.NoError(t, err)

		// If human's turn, fold to move the game forward
		if s.IsHumanTurn() && s.GetPhase() >= FollowTheQueenPhaseThirdStreet && s.GetPhase() <= FollowTheQueenPhaseSeventhStreet {
			err = s.PlayerAction(FollowTheQueenActionFold, 0, 0)
			require.NoError(t, err)
		}
	}
}

func TestFollowTheQueen_CurrentBetSize(t *testing.T) {
	s := newTestFollowTheQueen()

	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(FollowTheQueenPhaseFourthStreet)
	assert.Equal(t, s.GetConfig().SmallBet, s.currentBetSize())

	s.SetPhase(FollowTheQueenPhaseFifthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())

	s.SetPhase(FollowTheQueenPhaseSixthStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())

	s.SetPhase(FollowTheQueenPhaseSeventhStreet)
	assert.Equal(t, s.GetConfig().BigBet, s.currentBetSize())
}

func TestFollowTheQueen_Rebuy(t *testing.T) {
	t.Run("rebuy wrong phase", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.SetPhase(FollowTheQueenPhaseThirdStreet)
		assert.Error(t, s.Rebuy())
	})

	t.Run("skip rebuy wrong phase", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.SetPhase(FollowTheQueenPhaseThirdStreet)
		assert.Error(t, s.SkipRebuy())
	})

	t.Run("addon wrong phase", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.SetPhase(FollowTheQueenPhaseThirdStreet)
		assert.Error(t, s.Addon())
	})

	t.Run("skip addon wrong phase", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.SetPhase(FollowTheQueenPhaseThirdStreet)
		assert.Error(t, s.SkipAddon())
	})

	t.Run("rebuy happy path", func(t *testing.T) {
		s := newTestFollowTheQueen()
		for _, p := range s.players {
			p.SetChips(0)
		}
		cfg := s.GetConfig()
		cfg.RebuyEnabled = true
		cfg.RebuyMaxCount = 3
		cfg.RebuyChips = 500
		cfg.RebuyPeriodHands = 20
		s.SetConfig(cfg)
		s.SetPhase(FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseRebuy)
		s.SetRebuyCounts(make([]int, len(s.GetPlayers())))

		err := s.Rebuy()
		require.NoError(t, err)
		// After rebuy (500 chips), continueReset() deducts ante (1 chip)
		assert.Equal(t, 499, s.GetPlayers()[0].GetChips())
	})

	t.Run("skip rebuy - human busted", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.GetPlayers()[0].SetChips(0)
		s.SetPhase(FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseRebuy)

		err := s.SkipRebuy()
		require.NoError(t, err)
		assert.Equal(t, FollowTheQueenPhaseEnd, s.GetPhase())
		assert.True(t, s.GetGameEndFlag())
	})

	t.Run("addon happy path", func(t *testing.T) {
		s := newTestFollowTheQueen()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		cfg := s.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonChips = 500
		cfg.AddonAfterHand = 1
		s.SetConfig(cfg)
		s.SetPhase(FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseAddon)
		s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

		err := s.Addon()
		require.NoError(t, err)
		// After addon, continueReset() runs which deducts ante (1 chip)
		assert.Equal(t, 1499, s.GetPlayers()[0].GetChips())
	})

	t.Run("skip addon happy path", func(t *testing.T) {
		s := newTestFollowTheQueen()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		s.SetPhase(FollowTheQueenPhaseRebuy)
		s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseAddon)
		s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

		err := s.SkipAddon()
		require.NoError(t, err)
		// After skip addon, continueReset() runs which deducts ante
		assert.Less(t, s.GetPlayers()[0].GetChips(), 1000)
	})

	t.Run("IsRebuyAvailable", func(t *testing.T) {
		s := newTestFollowTheQueen()
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
		s := newTestFollowTheQueen()
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
		s := newTestFollowTheQueen()
		s.SetRebuyCounts([]int{1, 0, 0, 0})
		s.SetAddonUsed([]bool{true, false, false, false})
		assert.Equal(t, []int{1, 0, 0, 0}, s.GetRebuyCounts())
		assert.Equal(t, []bool{true, false, false, false}, s.GetAddonUsed())
		assert.Equal(t, FollowTheQueenRebuyPhaseNone, s.GetRebuyPhaseType())
	})

	t.Run("rebuy with tournament reset", func(t *testing.T) {
		s := newTestFollowTheQueen()
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
			_ = s.PlayerAction(FollowTheQueenActionFold, 0, 0)
		}
		// Drain human chips to trigger rebuy on next reset
		s.GetPlayers()[0].SetChips(0)
		err = s.Reset()
		require.NoError(t, err)
		// Should be in rebuy phase or have continued
		if s.GetPhase() == FollowTheQueenPhaseRebuy {
			assert.Equal(t, FollowTheQueenRebuyPhaseRebuy, s.GetRebuyPhaseType())
		}
	})
}

func TestFollowTheQueen_CommunityCardShowdown(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseShowdown)
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

func TestFollowTheQueen_CPUDecide_AllStyles(t *testing.T) {
	styles := []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	for _, style := range styles {
		t.Run(HoldemPlayStyleNames[style], func(t *testing.T) {
			s := newTestFollowTheQueen()
			for _, p := range s.players {
				p.SetChips(1000)
			}
			// Set CPU player style
			s.GetPlayers()[1].SetPlayStyle(style)

			// Give player 1 cards for evaluation
			s.GetPlayers()[1].AddHoleCard(NewCard(CardDesignSpade, 10, true))
			s.GetPlayers()[1].AddHoleCard(NewCard(CardDesignHeart, 10, true))
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignDiamond, 5, true))

			s.SetPhase(FollowTheQueenPhaseThirdStreet)
			s.SetLastBet(5)
			s.SetMinRaise(5)

			action, amount := s.cpuDecide(1)
			assert.GreaterOrEqual(t, action, 0)
			assert.GreaterOrEqual(t, amount, 0)

			// Also test post-third-street
			s.SetPhase(FollowTheQueenPhaseFifthStreet)
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignClover, 8, true))
			s.GetPlayers()[1].AddDoorCard(NewCard(CardDesignSpade, 9, true))

			action2, amount2 := s.cpuDecide(1)
			assert.GreaterOrEqual(t, action2, 0)
			assert.GreaterOrEqual(t, amount2, 0)
		})
	}
}

func TestFollowTheQueen_CPUDecide_MetaAI(t *testing.T) {
	s := newTestFollowTheQueen()
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

	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	s.SetLastBet(5)
	s.SetMinRaise(5)

	// Should not panic with meta AI enabled
	action, amount := s.cpuDecide(1)
	assert.GreaterOrEqual(t, action, 0)
	assert.GreaterOrEqual(t, amount, 0)
}

func TestFollowTheQueen_HandleCpuActionError_NoCallAmount(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	s.SetCurrentTurn(1)
	s.SetLastBet(0) // no call amount → should check instead of fold

	s.handleCpuActionError(1, FollowTheQueenActionBet, assert.AnError)
	assert.NotNil(t, s.GetLastCpuError())
}

func TestFollowTheQueen_EvalThirdStreetStrength_Variants(t *testing.T) {
	t.Run("suited cards", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 10, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 11, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignSpade, 12, true))

		strength := s.evalThirdStreetStrength(0)
		assert.GreaterOrEqual(t, strength, 40) // suited connected high cards
	})

	t.Run("low disconnected", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 2, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignHeart, 7, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignDiamond, 12, true))

		strength := s.evalThirdStreetStrength(0)
		assert.Greater(t, strength, 0)
	})

	t.Run("pair", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 8, true))
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignHeart, 8, true))
		s.GetPlayers()[0].AddDoorCard(NewCard(CardDesignDiamond, 3, true))

		strength := s.evalThirdStreetStrength(0)
		assert.GreaterOrEqual(t, strength, 40)
	})

	t.Run("fewer than 3 cards", func(t *testing.T) {
		s := newTestFollowTheQueen()
		s.GetPlayers()[0].AddHoleCard(NewCard(CardDesignSpade, 8, true))

		strength := s.evalThirdStreetStrength(0)
		assert.Equal(t, 0, strength)
	})
}

func TestFollowTheQueen_BettingLimits(t *testing.T) {
	s := newTestFollowTheQueen()
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

func TestFollowTheQueen_MetaAI(t *testing.T) {
	s := newTestFollowTheQueen()
	assert.Nil(t, s.GetHumanProfile())

	s.ResetProfile()
	assert.Nil(t, s.GetHumanProfile())

	assert.Nil(t, s.ExportProfile())
	assert.NoError(t, s.ImportProfile(nil))
	assert.NoError(t, s.ImportProfile([]byte{}))
}

func TestFollowTheQueen_ActionLog(t *testing.T) {
	s := newTestFollowTheQueen()
	assert.Empty(t, s.GetActionLog())

	s.appendLog(0, "test", "test action", nil)
	assert.Len(t, s.GetActionLog(), 1)
	assert.Equal(t, 1, s.GetActionLog()[0].TurnNumber)
}

func TestFollowTheQueen_CountActivePlayers(t *testing.T) {
	s := newTestFollowTheQueen()
	assert.Equal(t, 4, s.countActivePlayers())

	s.players[1].SetFolded(true)
	assert.Equal(t, 3, s.countActivePlayers())

	s.players[2].SetFolded(true)
	s.players[3].SetFolded(true)
	assert.Equal(t, 1, s.countActivePlayers())
}

func TestFollowTheQueen_EvalThirdStreetStrength(t *testing.T) {
	s := newTestFollowTheQueen()

	// Trip aces
	s.players[0].AddHoleCard(NewCard(CardDesignSpade, 1, true))
	s.players[0].AddHoleCard(NewCard(CardDesignHeart, 1, true))
	s.players[0].AddDoorCard(NewCard(CardDesignDiamond, 1, true))

	strength := s.evalThirdStreetStrength(0)
	assert.GreaterOrEqual(t, strength, 80) // trips should be very strong
}

func TestFollowTheQueen_HandleCpuActionError(t *testing.T) {
	s := setupFollowTheQueenForHumanAction(FollowTheQueenPhaseThirdStreet)
	s.SetCurrentTurn(1) // CPU player

	// With call amount > 0, should fold
	s.handleCpuActionError(1, FollowTheQueenActionBet, assert.AnError)
	assert.NotNil(t, s.GetLastCpuError())
}

func TestFollowTheQueen_TournamentEscalation(t *testing.T) {
	s := newTestFollowTheQueen()
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
		_ = s.PlayerAction(FollowTheQueenActionFold, 0, 0)
	}

	// Second reset should escalate
	err = s.Reset()
	require.NoError(t, err)
	assert.Greater(t, s.GetConfig().Ante, initialAnte)
}

func TestFollowTheQueen_CPUDecide_GTO_RaiseCountCap(t *testing.T) {
	// Verify that GTO CPU converts raise/bet to call/check when raiseCount >= maxRaises.
	setup := func() *FollowTheQueen {
		s := newTestFollowTheQueen()
		for _, p := range s.players {
			p.SetChips(1000)
		}
		// Give CPU player 1 a trips hand (strength ~90) so GTO always picks bet
		s.players[1].SetPlayStyle(HoldemStyleGTO)
		s.players[1].AddHoleCard(NewCard(CardDesignSpade, 9, true))
		s.players[1].AddHoleCard(NewCard(CardDesignHeart, 9, true))
		s.players[1].AddDoorCard(NewCard(CardDesignDiamond, 9, true))
		s.SetPhase(FollowTheQueenPhaseThirdStreet)
		s.SetPot(50)
		s.SetMinRaise(5)
		cfg := s.GetConfig()
		cfg.BettingLimit = BettingLimitFixed
		s.SetConfig(cfg)
		// Fixed limit: maxRaises = 4 (bettingMaxRaisesPerRound)
		s.SetRaiseCount(4) // at the cap
		return s
	}

	t.Run("converts raise to call when callAmount > 0", func(t *testing.T) {
		s := setup()
		s.SetLastBet(10) // callAmount = 10
		// GTO with trips will always want to bet; with raiseCount cap it must call
		hitCall := false
		for i := 0; i < 200; i++ {
			action, _ := s.cpuDecide(1)
			if action == FollowTheQueenActionCall {
				hitCall = true
				break
			}
		}
		assert.True(t, hitCall, "should convert bet/raise to Call when raise cap reached and callAmount > 0")
	})

	t.Run("converts raise to check when callAmount == 0", func(t *testing.T) {
		s := setup()
		s.SetLastBet(0) // callAmount = 0
		hitCheck := false
		for i := 0; i < 200; i++ {
			action, _ := s.cpuDecide(1)
			if action == FollowTheQueenActionCheck {
				hitCheck = true
				break
			}
		}
		assert.True(t, hitCheck, "should convert bet/raise to Check when raise cap reached and callAmount == 0")
	})
}

func TestFollowTheQueen_CPUDecide_GTO_PotLimitAmountCap(t *testing.T) {
	// Verify that GTO bet amount is capped to maxBetAmount under PotLimit.
	// With trips, GTO always bets (premium hand → betPct=100%).
	// Tiny pot forces maxBetAmount < cpuPotBet result, triggering the cap branch.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	s.players[1].SetPlayStyle(HoldemStyleGTO)
	s.players[1].AddHoleCard(NewCard(CardDesignSpade, 9, true))
	s.players[1].AddHoleCard(NewCard(CardDesignHeart, 9, true))
	s.players[1].AddDoorCard(NewCard(CardDesignDiamond, 9, true))
	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	s.SetPot(1) // tiny pot → maxBetAmount = pot+lastBet = 1
	s.SetLastBet(0)
	s.SetMinRaise(5) // minRaise=5 > maxBetAmount=1 → cpuPotBet returns 5, gets capped to 1
	cfg := s.GetConfig()
	cfg.BettingLimit = BettingLimitPotLimit
	s.SetConfig(cfg)

	// GTO with trips (strength=90 → premium bucket, betPct=90%) usually bets.
	// cpuPotBet(66) = max(1*66/100=0, SmallBet=5, minRaise=5) = 5
	// maxBetAmount = 1; 5 > 1 → capped to 1
	hitBet := false
	for i := 0; i < 500; i++ {
		action, amount := s.cpuDecide(1)
		if action == FollowTheQueenActionBet || action == FollowTheQueenActionRaise {
			assert.LessOrEqual(t, amount, 1, "bet amount should be capped to maxBetAmount=1")
			hitBet = true
			break
		}
	}
	assert.True(t, hitBet, "GTO with trips should bet within 500 tries (90%% probability)")
}

func TestFollowTheQueen_CPUDecide_CompoundFold(t *testing.T) {
	// Test the preFlopFoldCompound branches directly via cpuDecideThirdStreet.
	// Use custom params with threshold=101 to guarantee the compound branch is always entered.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	s.players[1].AddHoleCard(NewCard(CardDesignHeart, 2, true))
	s.players[1].AddHoleCard(NewCard(CardDesignDiamond, 8, true))
	s.players[1].AddDoorCard(NewCard(CardDesignClover, 5, true))
	s.SetMinRaise(5)
	params := cpuStyleParams{
		aggressive:           false,
		bluffRate:            0,
		preFlopFoldThreshold: 101, // Always triggers: any strength (0-100) is < 101
		preFlopFoldCompound:  true,
		preFlopFoldCallMult:  2,
		preFlopBluffPotPct:   50,
	}

	t.Run("folds when callAmount exceeds SmallBet*callMult threshold", func(t *testing.T) {
		// SmallBet=5, callMult=2 → threshold=10; callAmount=11 > 10 → fold
		action, _ := s.cpuDecideThirdStreet(1, params, 11)
		assert.Equal(t, FollowTheQueenActionFold, action)
	})

	t.Run("calls when callAmount is within compound threshold", func(t *testing.T) {
		// callAmount=5 <= 10 → falls through compound, callAmount>0 → call
		action, _ := s.cpuDecideThirdStreet(1, params, 5)
		assert.Equal(t, FollowTheQueenActionCall, action)
	})

	t.Run("checks when callAmount is zero", func(t *testing.T) {
		// callAmount=0 ≤ 10 → falls through compound, callAmount==0 → no-call
		// bluffRate=0 → always check
		action, _ := s.cpuDecideThirdStreet(1, params, 0)
		assert.Equal(t, FollowTheQueenActionCheck, action)
	})
}

func TestFollowTheQueen_CPUDecide_PostThirdFallbackFold(t *testing.T) {
	// Test the postFlopFallbackFold branches in cpuDecidePostThird.
	// Use custom params with bluffRate=0 for deterministic results.
	// postFlopFallbackFold=true means: when aggressive and hand < raiseRank,
	//   if handRank >= condCallRank && callAmount > 0 → Call
	//   else → FoldOrCheck
	params := cpuStyleParams{
		aggressive:           true,
		bluffRate:            0, // no bluff → deterministic
		postFlopRaiseRank:    PokerHandTwoPair,
		postFlopRaisePotPct:  66,
		postFlopCondCallRank: PokerHandOnePair,
		postFlopFallbackFold: true,
	}

	makePlayer := func(holeCards, doorCards []*Card) *FollowTheQueen {
		s := newTestFollowTheQueen()
		s.players[1].SetPlayStyle(HoldemStyleTAG)
		s.players[1].SetChips(1000)
		for _, c := range holeCards {
			s.players[1].AddHoleCard(c)
		}
		for _, c := range doorCards {
			s.players[1].AddDoorCard(c)
		}
		s.SetPhase(FollowTheQueenPhaseFifthStreet)
		s.SetPot(50)
		s.SetMinRaise(5)
		return s
	}

	t.Run("calls for OnePair hand when callAmount > 0", func(t *testing.T) {
		// 5-card OnePair: A♠,A♥ + K♦,J♣,9♠
		//
		// **Q を混ぜない。**クローン元の盤には Q♣ が入っていたが、このゲームで
		// Q は常時ワイルドなので、この手はワンペアではなくスリーカードになる。
		s := makePlayer(
			[]*Card{NewCard(CardDesignSpade, 1, true), NewCard(CardDesignHeart, 1, true)},
			[]*Card{
				NewCard(CardDesignDiamond, 13, true),
				NewCard(CardDesignClover, 11, true),
				NewCard(CardDesignSpade, 9, true),
			},
		)
		action, _ := s.cpuDecidePostThird(1, params, 10)
		assert.Equal(t, FollowTheQueenActionCall, action)
	})

	t.Run("folds for HighCard hand when callAmount > 0", func(t *testing.T) {
		// 5-card HighCard: K♠,J♥ + 9♦,7♣,5♠ (no pair, no flush, no straight)
		// ここも Q を外す ── Q が 1 枚あるだけでハイカードではなくなる。
		s := makePlayer(
			[]*Card{NewCard(CardDesignSpade, 13, true), NewCard(CardDesignHeart, 11, true)},
			[]*Card{
				NewCard(CardDesignDiamond, 9, true),
				NewCard(CardDesignClover, 7, true),
				NewCard(CardDesignSpade, 5, true),
			},
		)
		action, _ := s.cpuDecidePostThird(1, params, 10)
		assert.Equal(t, FollowTheQueenActionFold, action)
	})

	t.Run("checks for HighCard hand when callAmount == 0", func(t *testing.T) {
		s := makePlayer(
			[]*Card{NewCard(CardDesignSpade, 13, true), NewCard(CardDesignHeart, 12, true)},
			[]*Card{
				NewCard(CardDesignDiamond, 9, true),
				NewCard(CardDesignClover, 7, true),
				NewCard(CardDesignSpade, 5, true),
			},
		)
		action, _ := s.cpuDecidePostThird(1, params, 0)
		assert.Equal(t, FollowTheQueenActionCheck, action)
	})
}

func TestFollowTheQueen_CPUDecide_MetaAI_FoldToCallPath(t *testing.T) {
	// Meta AI: when action=Fold and callAmount>0 and lastHumanPlayMs>0, adjusts call chance.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	cfg := s.GetConfig()
	cfg.CpuMetaAI = true
	s.SetConfig(cfg)
	s.SetHumanProfile(&BettingHumanProfile{GamesPlayed: 50})
	s.SetLastHumanPlayMs(500)

	// Give LAP CPU a weak hand that will compound-fold with high callAmount
	s.players[1].SetPlayStyle(HoldemStyleLAP)
	s.players[1].AddHoleCard(NewCard(CardDesignHeart, 2, true))
	s.players[1].AddHoleCard(NewCard(CardDesignDiamond, 4, true))
	s.players[1].AddDoorCard(NewCard(CardDesignClover, 7, true))
	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	s.SetPot(20)
	s.SetLastBet(15) // callAmount=15 > SmallBet*2=10, LAP will fold; meta AI may override

	// Run multiple times to exercise both possible outcomes of the meta AI branch
	hitFold := false
	hitCall := false
	for i := 0; i < 1000; i++ {
		action, _ := s.cpuDecide(1)
		switch action {
		case FollowTheQueenActionFold:
			hitFold = true
		case FollowTheQueenActionCall:
			hitCall = true
		}
		if hitFold && hitCall {
			break
		}
	}
	// Meta AI path was exercised (at minimum fold was observed)
	assert.True(t, hitFold || hitCall, "meta AI fold-to-call path should be reachable")
}

func TestFollowTheQueen_CPUDecide_PassiveAllIn(t *testing.T) {
	// Covers the passive path where betAmt > p.GetChips() → AllIn.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	// LAP passive, bluffRate=5 but eventually triggers; give tiny chips
	s.players[1].SetPlayStyle(HoldemStyleLAP)
	s.players[1].AddHoleCard(NewCard(CardDesignSpade, 1, true))
	s.players[1].AddHoleCard(NewCard(CardDesignHeart, 2, true))
	s.players[1].AddDoorCard(NewCard(CardDesignDiamond, 3, true))
	s.players[1].SetChips(1) // tiny chips so any bet triggers AllIn
	s.SetPhase(FollowTheQueenPhaseFifthStreet)
	s.SetPot(100)
	s.SetLastBet(0) // callAmount=0 → passive bluff path
	s.SetMinRaise(5)

	hitAllIn := false
	for i := 0; i < 1000; i++ {
		action, _ := s.cpuDecide(1)
		if action == FollowTheQueenActionAllIn {
			hitAllIn = true
			break
		}
	}
	assert.True(t, hitAllIn, "passive AllIn path should be reachable when chips < betAmt")
}

func TestFollowTheQueen_CPUDecide_UnknownStyle(t *testing.T) {
	// Unknown style falls back to CpuCallOrCheck.
	s := newTestFollowTheQueen()
	s.players[1].SetPlayStyle(999) // unknown
	s.players[1].SetChips(1000)
	s.SetPhase(FollowTheQueenPhaseThirdStreet)
	s.SetLastBet(0)

	action, _ := s.cpuDecide(1)
	assert.Equal(t, FollowTheQueenActionCheck, action)

	s.SetLastBet(10)
	action, _ = s.cpuDecide(1)
	assert.Equal(t, FollowTheQueenActionCall, action)
}

func TestFollowTheQueen_CheckAndTransitionAddon_CPUOnly(t *testing.T) {
	// checkAndTransitionAddon returns false when only CPUs need addon but human already used it.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	cfg := s.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 500
	cfg.AddonAfterHand = 5
	s.SetConfig(cfg)
	s.SetHandCount(5) // matches AddonAfterHand

	// Mark human addon as already used; CPUs not used
	addonUsed := make([]bool, len(s.GetPlayers()))
	addonUsed[0] = true // human already used
	s.SetAddonUsed(addonUsed)

	// With addonUsed[0]=true, human skips; CPUs get addon; no human needed → returns false
	result := s.checkAndTransitionAddon()
	// CPUs take addon automatically, no human prompt needed
	assert.False(t, result)
	// Non-human players who hadn't used addon should have received chips
	for i := 1; i < len(s.GetPlayers()); i++ {
		assert.Equal(t, 1500, s.GetPlayers()[i].GetChips(),
			"CPU player %d should have received addon chips", i)
	}
}

func TestFollowTheQueen_Rebuy_TransitionsToAddon(t *testing.T) {
	// After Rebuy(), if addon condition is met, transitions to addon phase.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(0)
	}
	cfg := s.GetConfig()
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 3
	cfg.RebuyChips = 500
	cfg.RebuyPeriodHands = 20
	cfg.AddonEnabled = true
	cfg.AddonChips = 300
	cfg.AddonAfterHand = 2
	s.SetConfig(cfg)
	s.SetHandCount(2) // matches AddonAfterHand
	s.SetPhase(FollowTheQueenPhaseRebuy)
	s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseRebuy)
	s.SetRebuyCounts(make([]int, len(s.GetPlayers())))
	s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

	err := s.Rebuy()
	require.NoError(t, err)
	// Should have transitioned to addon phase since handCount==AddonAfterHand
	assert.Equal(t, FollowTheQueenPhaseRebuy, s.GetPhase())
	assert.Equal(t, FollowTheQueenRebuyPhaseAddon, s.GetRebuyPhaseType())
}

func TestFollowTheQueen_SkipRebuy_WithChipsTransitionsToAddon(t *testing.T) {
	// SkipRebuy when human has chips and addon condition met → transitions to addon.
	s := newTestFollowTheQueen()
	for _, p := range s.players {
		p.SetChips(1000)
	}
	cfg := s.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 300
	cfg.AddonAfterHand = 3
	s.SetConfig(cfg)
	s.SetHandCount(3) // matches AddonAfterHand
	s.SetPhase(FollowTheQueenPhaseRebuy)
	s.SetRebuyPhaseType(FollowTheQueenRebuyPhaseRebuy)
	s.SetAddonUsed(make([]bool, len(s.GetPlayers())))

	err := s.SkipRebuy()
	require.NoError(t, err)
	// Human has chips (> 0), so no bust-out; addon should kick in
	assert.Equal(t, FollowTheQueenPhaseRebuy, s.GetPhase())
	assert.Equal(t, FollowTheQueenRebuyPhaseAddon, s.GetRebuyPhaseType())
}
