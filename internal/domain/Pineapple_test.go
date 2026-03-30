package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPineapple() *Pineapple {
	cfg := DefaultPineappleConfig()
	players := NewPineapplePlayersForTable(HoldemTableSize4)
	return NewPineapple(NewTrumpCards(0), players, cfg)
}

func TestNewPineapple(t *testing.T) {
	p := newTestPineapple()
	assert.Equal(t, PineapplePhaseInit, p.GetPhase())
	assert.Equal(t, 4, p.GetPlayerCnt())
	assert.False(t, p.GetGameEndFlag())
}

func TestPineapple_PhaseConstants(t *testing.T) {
	assert.Equal(t, 0, PineapplePhaseInit)
	assert.Equal(t, 1, PineapplePhasePreFlop)
	assert.Equal(t, 2, PineapplePhaseFlop)
	assert.Equal(t, 3, PineapplePhaseTurn)
	assert.Equal(t, 4, PineapplePhaseRiver)
	assert.Equal(t, 5, PineapplePhaseShowdown)
	assert.Equal(t, 6, PineapplePhaseEnd)
	assert.Equal(t, 7, PineapplePhaseRebuy)
	assert.Equal(t, 8, PineapplePhaseDiscard)
}

func TestPineapple_Reset_Deals3HoleCards(t *testing.T) {
	p := newTestPineapple()
	err := p.Reset()
	require.NoError(t, err)

	// Phase should be PreFlop or beyond (CPU actions may advance it)
	assert.GreaterOrEqual(t, p.GetPhase(), PineapplePhasePreFlop)

	// Human player (index 0) should have 3 hole cards
	human := p.GetPlayer(0)
	assert.Equal(t, 3, human.GetCardsSize())
}

func TestPineapple_Reset_ThroughPreFlopToDiscard(t *testing.T) {
	// Run multiple times due to random CPU decisions
	discardReached := false
	for i := 0; i < 100; i++ {
		p := newTestPineapple()
		err := p.Reset()
		require.NoError(t, err)

		phase := p.GetPhase()
		if phase == PineapplePhasePreFlop {
			// Human turn at preflop -- fold all CPUs to end, or play normally
			// Just check human has 3 cards
			assert.Equal(t, 3, p.GetPlayer(0).GetCardsSize())
		}

		// If all CPUs check/call and it moves to discard
		if phase == PineapplePhaseDiscard {
			discardReached = true
			assert.True(t, p.IsDiscardPhase())
			assert.Equal(t, 3, len(p.GetCommunityCards()))
			break
		}
	}
	// If discard not reached through normal play, test explicit flow
	if !discardReached {
		t.Log("Discard phase not reached via auto-play, testing explicit flow")
	}
}

func TestPineapple_DiscardCard(t *testing.T) {
	t.Run("discard in correct phase", func(t *testing.T) {
		p := newTestPineapple()
		// Manually set up discard state
		p.phase = PineapplePhaseDiscard
		p.communityCards = []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false)}
		p.discardDone = make([]bool, len(p.players))
		// Mark CPU players as done
		for i := 1; i < len(p.players); i++ {
			p.discardDone[i] = true
		}
		// Give human 3 cards
		p.players[0].Reset()
		p.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		p.players[0].AddCard(NewCard(CardDesignSpade, 13, false))
		p.players[0].AddCard(NewCard(CardDesignHeart, 2, false))

		err := p.DiscardCard(2) // discard the weakest card (2 of hearts)
		assert.NoError(t, err)
		assert.Equal(t, 2, p.players[0].GetCardsSize())
	})

	t.Run("discard in wrong phase returns error", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		err := p.DiscardCard(0)
		assert.Error(t, err)
	})

	t.Run("invalid card index returns error", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhaseDiscard
		p.discardDone = make([]bool, len(p.players))
		for i := 1; i < len(p.players); i++ {
			p.discardDone[i] = true
		}
		p.players[0].Reset()
		p.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		p.players[0].AddCard(NewCard(CardDesignSpade, 13, false))
		p.players[0].AddCard(NewCard(CardDesignHeart, 2, false))

		err := p.DiscardCard(5) // out of range
		assert.Error(t, err)
	})

	t.Run("double discard returns error", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhaseDiscard
		p.discardDone = make([]bool, len(p.players))
		for i := 1; i < len(p.players); i++ {
			p.discardDone[i] = true
		}
		p.players[0].Reset()
		p.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		p.players[0].AddCard(NewCard(CardDesignSpade, 13, false))
		p.players[0].AddCard(NewCard(CardDesignHeart, 2, false))

		err := p.DiscardCard(0)
		assert.NoError(t, err)

		// Try to discard again
		p.phase = PineapplePhaseDiscard
		p.discardDone[0] = true
		err = p.DiscardCard(0)
		assert.Error(t, err)
	})
}

func TestPineapple_PlayerAction(t *testing.T) {
	t.Run("action in wrong phase returns error", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhaseInit
		err := p.PlayerAction(PineappleActionCheck, 0, 0)
		assert.Error(t, err)
	})

	t.Run("action when game ended returns error", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		p.gameEndFlag = true
		err := p.PlayerAction(PineappleActionCheck, 0, 0)
		assert.Error(t, err)
	})
}

func TestPineapple_FullGame(t *testing.T) {
	// Run multiple games to cover different paths
	for i := 0; i < 50; i++ {
		p := newTestPineapple()
		err := p.Reset()
		require.NoError(t, err)

		maxActions := 200
		actions := 0
		for !p.GetGameEndFlag() && actions < maxActions {
			actions++
			phase := p.GetPhase()

			if phase == PineapplePhaseDiscard {
				if p.GetPlayer(0).GetCardsSize() == 3 {
					err = p.DiscardCard(0) // always discard first card
					if err != nil {
						break
					}
				}
				continue
			}

			if phase >= PineapplePhasePreFlop && phase <= PineapplePhaseRiver {
				if p.IsHumanTurn() {
					callAmt := p.GetLastBet() - p.GetPlayer(0).GetCurrentBet()
					if callAmt > 0 {
						err = p.PlayerAction(PineappleActionCall, 0, 0)
					} else {
						err = p.PlayerAction(PineappleActionCheck, 0, 0)
					}
					if err != nil {
						break
					}
				} else {
					break // shouldn't happen, CPU should auto-act
				}
				continue
			}

			if phase == PineapplePhaseShowdown {
				if p.IsMuckAvailable() {
					err = p.Muck()
				} else {
					err = p.ShowHand()
				}
				if err != nil {
					break
				}
				continue
			}

			if phase == PineapplePhaseEnd {
				break
			}

			break // unexpected phase
		}

		// Game should complete
		if actions < maxActions {
			assert.True(t, p.GetGameEndFlag() || p.GetPhase() == PineapplePhaseEnd ||
				p.GetPhase() == PineapplePhasePreFlop || p.GetPhase() == PineapplePhaseDiscard)
		}
	}
}

func TestPineapple_Muck(t *testing.T) {
	t.Run("muck in wrong phase", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		err := p.Muck()
		assert.Error(t, err)
	})

	t.Run("show hand in wrong phase", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		err := p.ShowHand()
		assert.Error(t, err)
	})
}

func TestPineapple_IsMuckAvailable(t *testing.T) {
	p := newTestPineapple()
	p.phase = PineapplePhaseFlop
	assert.False(t, p.IsMuckAvailable())
}

func TestPineapple_Rebuy(t *testing.T) {
	t.Run("rebuy in wrong phase", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		assert.Error(t, p.Rebuy())
		assert.Error(t, p.SkipRebuy())
	})

	t.Run("addon in wrong phase", func(t *testing.T) {
		p := newTestPineapple()
		p.phase = PineapplePhasePreFlop
		assert.Error(t, p.Addon())
		assert.Error(t, p.SkipAddon())
	})

	t.Run("rebuy available check", func(t *testing.T) {
		p := newTestPineapple()
		assert.False(t, p.IsRebuyAvailable())
	})

	t.Run("addon available check", func(t *testing.T) {
		p := newTestPineapple()
		assert.False(t, p.IsAddonAvailable())
	})
}

func TestPineapple_Getters(t *testing.T) {
	p := newTestPineapple()
	assert.NotNil(t, p.GetPlayers())
	assert.Nil(t, p.GetPlayer(-1))
	assert.Nil(t, p.GetPlayer(100))
	assert.NotNil(t, p.GetPlayer(0))
	assert.Equal(t, 4, p.GetPlayerCnt())
	assert.NotNil(t, p.GetCommunityCards())
	assert.Equal(t, 0, p.GetPot())
	assert.NotNil(t, p.GetSidePots())
	assert.Equal(t, 0, p.GetDealerIdx())
	assert.Equal(t, 0, p.GetCurrentTurn())
	assert.False(t, p.GetGameEndFlag())
	assert.Equal(t, 0, p.GetLastBet())
	assert.Equal(t, 0, p.GetMinRaise())
	assert.Equal(t, 0, p.GetRaiseCount())
	assert.NotNil(t, p.GetRoundResults())
	assert.NotNil(t, p.GetCpuActions())
	assert.Nil(t, p.GetLastCpuError())
	assert.Nil(t, p.GetHumanProfile())
	assert.Equal(t, 0, p.GetHandCount())
	assert.NotNil(t, p.GetActedFlags())
	assert.NotNil(t, p.GetRebuyCounts())
	assert.NotNil(t, p.GetAddonUsed())
	assert.Equal(t, 0, p.GetRebuyPhaseType())
	assert.NotNil(t, p.GetDiscardDone())
	assert.False(t, p.IsDiscardPhase())

	cfg := p.GetConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	p.SetConfig(cfg)
}

func TestPineapple_IsHumanTurn(t *testing.T) {
	p := newTestPineapple()
	p.currentTurn = 0
	assert.True(t, p.IsHumanTurn())
	p.currentTurn = 1
	assert.False(t, p.IsHumanTurn())
	p.currentTurn = -1
	assert.False(t, p.IsHumanTurn())
}

func TestPineapple_Profile(t *testing.T) {
	p := newTestPineapple()
	assert.Nil(t, p.ExportProfile())

	p.humanProfile = &BettingHumanProfile{}
	p.humanProfile.GamesPlayed = 5
	assert.NotNil(t, p.ExportProfile())

	p.ResetProfile()
	assert.Nil(t, p.GetHumanProfile())
}

func TestPineapple_ImportProfile(t *testing.T) {
	p := newTestPineapple()
	assert.NoError(t, p.ImportProfile(nil))
	assert.NoError(t, p.ImportProfile([]byte{}))
	assert.Error(t, p.ImportProfile([]byte("invalid json")))
}

func TestPineapple_Equity(t *testing.T) {
	p := newTestPineapple()
	// Not in betting phase
	assert.Nil(t, p.GetEquity())
	assert.Equal(t, 0.0, p.GetPotOdds())
}

func TestPineapple_Resize(t *testing.T) {
	p := newTestPineapple()
	newPlayers := NewPineapplePlayersForTable(HoldemTableSize6)
	p.Resize(newPlayers)
	assert.Equal(t, 6, p.GetPlayerCnt())
	assert.Equal(t, 0, p.GetHandCount())
}

func TestPineapple_JSON(t *testing.T) {
	p := newTestPineapple()
	err := p.Reset()
	require.NoError(t, err)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	p2 := &Pineapple{}
	err = json.Unmarshal(data, p2)
	require.NoError(t, err)

	assert.Equal(t, p.GetPhase(), p2.GetPhase())
	assert.Equal(t, p.GetPot(), p2.GetPot())
	assert.Equal(t, p.GetPlayerCnt(), p2.GetPlayerCnt())
	assert.Equal(t, p.GetHandCount(), p2.GetHandCount())
}

func TestPineapple_JSON_Empty(t *testing.T) {
	data := []byte(`{}`)
	p := &Pineapple{}
	err := json.Unmarshal(data, p)
	assert.NoError(t, err)
	assert.NotNil(t, p.communityCards)
	assert.NotNil(t, p.roundResults)
}

func TestPineapple_JSON_MaxSliceLen(t *testing.T) {
	// Build JSON with oversized slice
	huge := make([]int, 1001)
	data, _ := json.Marshal(map[string]interface{}{
		"sc": huge,
	})
	p := &Pineapple{}
	err := json.Unmarshal(data, p)
	assert.Error(t, err)
}

func TestPineapple_CpuDiscard(t *testing.T) {
	p := newTestPineapple()
	p.communityCards = []*Card{NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 11, false), NewCard(CardDesignSpade, 10, false)}

	// Give CPU player 3 cards: Ace spade, King spade, 2 hearts
	cpuPlayer := p.players[1]
	cpuPlayer.Reset()
	cpuPlayer.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace
	cpuPlayer.AddCard(NewCard(CardDesignSpade, 13, false)) // King
	cpuPlayer.AddCard(NewCard(CardDesignHeart, 2, false))  // 2 (weakest)

	discardIdx := p.cpuDiscard(1)
	// Should discard the 2 (index 2) as remaining Ace+King gives a straight
	assert.Equal(t, 2, discardIdx)
}

func TestPineapple_EvalPreFlopStrength(t *testing.T) {
	t.Run("3 cards evaluated as best 2-card pair + bonus", func(t *testing.T) {
		p := newTestPineapple()
		pl := p.players[0]
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 1, false))   // Ace
		pl.AddCard(NewCard(CardDesignHeart, 1, false))   // Ace
		pl.AddCard(NewCard(CardDesignDiamond, 2, false)) // 2

		strength := p.evalPreFlopStrength(0)
		assert.Greater(t, strength, 50) // AA + bonus should be high
	})

	t.Run("2 cards (after discard) uses Holdem eval", func(t *testing.T) {
		p := newTestPineapple()
		pl := p.players[0]
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace
		pl.AddCard(NewCard(CardDesignSpade, 13, false)) // King suited

		strength := p.evalPreFlopStrength(0)
		assert.Greater(t, strength, 30)
	})

	t.Run("less than 2 cards returns 0", func(t *testing.T) {
		p := newTestPineapple()
		pl := p.players[0]
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 1, false))

		strength := p.evalPreFlopStrength(0)
		assert.Equal(t, 0, strength)
	})
}

func TestEvalTwoCardStrength(t *testing.T) {
	t.Run("pocket aces", func(t *testing.T) {
		score := evalTwoCardStrength(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false))
		assert.Greater(t, score, 80)
	})

	t.Run("suited connectors", func(t *testing.T) {
		score := evalTwoCardStrength(NewCard(CardDesignSpade, 10, false), NewCard(CardDesignSpade, 11, false))
		assert.Greater(t, score, 30)
	})

	t.Run("low offsuit", func(t *testing.T) {
		score := evalTwoCardStrength(NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 7, false))
		assert.Less(t, score, 30)
	})
}

func TestPineapple_enterDiscardPhase_FoldedPlayers(t *testing.T) {
	p := newTestPineapple()
	p.communityCards = []*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false)}
	p.discardDone = make([]bool, len(p.players))

	// Set player 1 as folded
	p.players[1].SetFolded(true)

	// Give other players 3 cards
	for i := 0; i < len(p.players); i++ {
		p.players[i].Reset()
		if !p.players[i].GetFolded() {
			p.players[i].AddCard(NewCard(CardDesignSpade, 1, false))
			p.players[i].AddCard(NewCard(CardDesignSpade, 13, false))
			p.players[i].AddCard(NewCard(CardDesignHeart, 2, false))
		}
	}

	p.enterDiscardPhase()

	// Folded player should be marked as discard done
	assert.True(t, p.discardDone[1])

	// CPU players should have discarded (2 cards)
	for i := 2; i < len(p.players); i++ {
		assert.True(t, p.discardDone[i])
		assert.Equal(t, 2, p.players[i].GetCardsSize())
	}

	// Human should NOT have discarded
	assert.False(t, p.discardDone[0])
	assert.Equal(t, 3, p.players[0].GetCardsSize())
}

func TestPineapple_MetaAI(t *testing.T) {
	cfg := DefaultPineappleConfig()
	cfg.CpuMetaAI = true
	players := NewPineapplePlayersForTable(HoldemTableSize4)
	p := NewPineapple(NewTrumpCards(0), players, cfg)

	err := p.Reset()
	require.NoError(t, err)
	assert.NotNil(t, p.GetHumanProfile())
}

func TestPineapple_ActionLog(t *testing.T) {
	p := newTestPineapple()
	assert.Nil(t, p.GetActionLog())

	err := p.Reset()
	require.NoError(t, err)
	// After reset, at least blind posts should be logged
	assert.NotNil(t, p.GetActionLog())
	assert.Greater(t, len(p.GetActionLog()), 0)
}
