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
