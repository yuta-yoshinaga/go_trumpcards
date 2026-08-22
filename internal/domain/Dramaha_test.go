package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestDramaha() *Dramaha {
	players := []*DramahaPlayer{
		NewDramahaPlayer(true, HoldemStyleTAG),
		NewDramahaPlayer(false, HoldemStyleTAG),
		NewDramahaPlayer(false, HoldemStyleLAP),
		NewDramahaPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultDramahaConfig()
	tc := NewTrumpCards(0)
	return NewDramaha(tc, players, cfg)
}

func setupDramahaForHumanAction(phase int) *Dramaha {
	o := newTestDramaha()
	for _, p := range o.players {
		p.SetChips(1000)
	}
	o.startingChips = []int{1000, 1000, 1000, 1000}
	o.SetPhase(phase)
	o.SetCurrentTurn(0)
	o.SetLastBet(0)
	o.SetMinRaise(10)
	o.SetPot(30)
	o.actedFlags = []bool{false, true, true, true}
	// Give every seat a full Dramaha hole of five cards.
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 13, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 11, false))
		p.AddCard(NewCard(CardDesignSpade, 10, false))
	}
	return o
}

func TestNewDramaha(t *testing.T) {
	o := newTestDramaha()
	assert.Equal(t, DramahaPhaseInit, o.GetPhase())
	assert.Equal(t, 4, o.GetPlayerCnt())
	assert.NotNil(t, o.GetCommunityCards())
	assert.Equal(t, 0, o.GetPot())
	assert.False(t, o.GetGameEndFlag())
}

func TestDramaha_Reset(t *testing.T) {
	o := newTestDramaha()
	_ = o.Reset()

	// After reset, each player should hold the full Dramaha hole of five.
	for i := 0; i < o.GetPlayerCnt(); i++ {
		p := o.GetPlayer(i)
		assert.Equal(t, 5, p.GetCardsSize(), "player %d should have 5 cards", i)
		assert.True(t, p.GetChips() > 0 || p.GetAllIn())
	}
	assert.True(t, o.GetPot() > 0)
}

// TestDramaha_Reset_DealsFiveHoleCards pins the deal width. Dramaha is not
// Omaha: the five hole cards *are* the draw hand, so a four-card deal would
// leave the draw side of the pot unevaluable.
func TestDramaha_Reset_DealsFiveHoleCards(t *testing.T) {
	assert.Equal(t, 5, DramahaHoleCards)

	o := newTestDramaha()
	_ = o.Reset()
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, DramahaHoleCards, o.GetPlayer(i).GetCardsSize(),
			"seat %d must be dealt exactly %d hole cards", i, DramahaHoleCards)
	}
	assert.Equal(t, DramahaHoleCards, o.GetHoleCardCount())
}

// TestDramaha_HoleCardCount_UnconditionalAfterRestore covers the divergence
// from the clone: Omaha defaulted an unset hole-card count to 4 and let Big O
// override it to 5. Dramaha has no such setting, so a table restored from a
// payload that never carried the field must still deal five.
func TestDramaha_HoleCardCount_UnconditionalAfterRestore(t *testing.T) {
	o := newTestDramaha()
	o.holeCards = 0 // as a payload written before the field existed would restore

	assert.Equal(t, DramahaHoleCards, o.holeCardCount(),
		"an unset holeCards field must not fall back to Omaha's 4")
	assert.Equal(t, DramahaHoleCards, o.GetHoleCardCount())

	_ = o.Reset()
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, DramahaHoleCards, o.GetPlayer(i).GetCardsSize(),
			"seat %d of a restored table must still be dealt %d cards", i, DramahaHoleCards)
	}
}

// TestDramaha_UnmarshalledTableDealsFiveHoleCards drives the same rule through
// real JSON rather than by poking the field: a serialised table whose payload
// omits "hcn" must come back dealing five.
func TestDramaha_UnmarshalledTableDealsFiveHoleCards(t *testing.T) {
	src := newTestDramaha()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	// Strip the hole-card field the way a payload written before the field
	// existed would arrive.
	stripped := strings.ReplaceAll(string(data), `,"hcn":5`, "")
	assert.NotContains(t, stripped, `"hcn"`)
	assert.NotEqual(t, string(data), stripped, "the fixture must actually drop the field")

	var restored Dramaha
	assert.NoError(t, restored.UnmarshalJSON([]byte(stripped)))
	assert.Equal(t, 0, restored.holeCards, "the restored table carries no hole-card setting")
	assert.Equal(t, DramahaHoleCards, restored.GetHoleCardCount())

	assert.NoError(t, restored.Reset())
	for i := 0; i < restored.GetPlayerCnt(); i++ {
		assert.Equal(t, DramahaHoleCards, restored.GetPlayer(i).GetCardsSize())
	}
}

func TestDramaha_Resize(t *testing.T) {
	o := newTestDramaha()
	assert.Equal(t, 4, o.GetPlayerCnt())

	newPlayers := make([]*DramahaPlayer, 6)
	newPlayers[0] = NewDramahaPlayer(true, HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = NewDramahaPlayer(false, HoldemStyleLAP)
	}
	o.Resize(newPlayers)
	assert.Equal(t, 6, o.GetPlayerCnt())
	assert.Equal(t, 0, o.GetHandCount())
	assert.Equal(t, 6, len(o.GetActedFlags()))

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 6, o.GetPlayerCnt())
	for i := 0; i < 6; i++ {
		assert.Equal(t, DramahaHoleCards, o.GetPlayer(i).GetCardsSize())
	}
}

func TestDramaha_PlayerAction_Fold(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.lastBet = 50
	o.players[0].SetCurrentBet(0)

	err := o.PlayerAction(DramahaActionFold, 0, 0)
	assert.NoError(t, err)
}

func TestDramaha_PlayerAction_Check(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(DramahaActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestDramaha_PlayerAction_Bet(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(DramahaActionBet, 20, 0)
	assert.NoError(t, err)
}

func TestDramaha_PlayerAction_GameEnded(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	o.gameEndFlag = true
	err := o.PlayerAction(DramahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestDramaha_PlayerAction_WrongPhase(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseShowdown)
	err := o.PlayerAction(DramahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestDramaha_PlayerAction_NotHumanTurn(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.currentTurn = 1 // CPU
	err := o.PlayerAction(DramahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestDramaha_Getters(t *testing.T) {
	o := newTestDramaha()
	_ = o.Reset()

	assert.NotNil(t, o.GetCommunityCards())
	assert.NotNil(t, o.GetSidePots())
	assert.Equal(t, 0, o.GetDealerIdx())
	// After reset, lastBet is BB (10) from blind posting
	assert.True(t, o.GetLastBet() >= 0)
	assert.NotNil(t, o.GetRoundResults())
	assert.NotNil(t, o.GetCpuActions())
	assert.Nil(t, o.GetLastCpuError())
	assert.Equal(t, 4, o.GetConfig().TableSize)
	assert.NotNil(t, o.GetActedFlags())
	assert.Equal(t, 1, o.GetHandCount())
	assert.NotNil(t, o.GetActionLog())
	assert.NotNil(t, o.GetRebuyCounts())
	assert.NotNil(t, o.GetAddonUsed())
	assert.Equal(t, DramahaRebuyPhaseNone, o.GetRebuyPhaseType())
}

func TestDramaha_GetPlayer_InvalidIndex(t *testing.T) {
	o := newTestDramaha()
	assert.Nil(t, o.GetPlayer(-1))
	assert.Nil(t, o.GetPlayer(100))
}

func TestDramaha_IsHumanTurn(t *testing.T) {
	o := newTestDramaha()
	_ = o.Reset()
	o.currentTurn = 0
	assert.True(t, o.IsHumanTurn())
	o.currentTurn = 1
	assert.False(t, o.IsHumanTurn())
}

func TestDramaha_SetConfig(t *testing.T) {
	o := newTestDramaha()
	cfg := DefaultDramahaConfig()
	cfg.BigBlind = 20
	o.SetConfig(cfg)
	assert.Equal(t, 20, o.GetConfig().BigBlind)
}

func TestDramaha_Muck(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.Muck())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		err := o.Muck()
		assert.NoError(t, err)
		assert.Equal(t, DramahaPhaseEnd, o.phase)
	})
}

func TestDramaha_ShowHand(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.ShowHand())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseShowdown
		err := o.ShowHand()
		assert.NoError(t, err)
		assert.Equal(t, DramahaPhaseEnd, o.phase)
	})
}

func TestDramaha_IsMuckAvailable(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.False(t, o.IsMuckAvailable())
	})
	t.Run("human lost", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		assert.True(t, o.IsMuckAvailable())
	})
	t.Run("human won", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseShowdown
		o.roundResults = []HoldemResult{
			{PlayerIdx: 0, WonAmount: 100},
		}
		assert.False(t, o.IsMuckAvailable())
	})
}

func TestDramaha_Rebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.Rebuy())
	})
	t.Run("wrong rebuy type", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseRebuy
		o.rebuyPhaseType = DramahaRebuyPhaseAddon
		assert.Error(t, o.Rebuy())
	})
}

func TestDramaha_SkipRebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.SkipRebuy())
	})
}

func TestDramaha_Addon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.Addon())
	})
	t.Run("wrong addon type", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseRebuy
		o.rebuyPhaseType = DramahaRebuyPhaseRebuy
		assert.Error(t, o.Addon())
	})
}

func TestDramaha_SkipAddon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestDramaha()
		o.phase = DramahaPhaseFlop
		assert.Error(t, o.SkipAddon())
	})
}

func TestDramaha_IsRebuyAvailable(t *testing.T) {
	t.Run("rebuy disabled", func(t *testing.T) {
		o := newTestDramaha()
		assert.False(t, o.IsRebuyAvailable())
	})
}

func TestDramaha_IsAddonAvailable(t *testing.T) {
	t.Run("addon disabled", func(t *testing.T) {
		o := newTestDramaha()
		assert.False(t, o.IsAddonAvailable())
	})
}

func TestDramaha_PhaseConstants(t *testing.T) {
	assert.Equal(t, 0, DramahaPhaseInit)
	assert.Equal(t, 1, DramahaPhasePreFlop)
	assert.Equal(t, 2, DramahaPhaseFlop)
	assert.Equal(t, 3, DramahaPhaseTurn)
	assert.Equal(t, 4, DramahaPhaseRiver)
	assert.Equal(t, 5, DramahaPhaseShowdown)
	assert.Equal(t, 6, DramahaPhaseEnd)
	assert.Equal(t, 7, DramahaPhaseRebuy)
	// The draw round is Dramaha's own phase and must not collide with any of
	// the Hold'em values it is aliased alongside -- a collision would route
	// the draw round into another game's branch.
	assert.Equal(t, 8, DramahaPhaseDraw)
	for _, taken := range []int{
		DramahaPhaseInit, DramahaPhasePreFlop, DramahaPhaseFlop, DramahaPhaseTurn,
		DramahaPhaseRiver, DramahaPhaseShowdown, DramahaPhaseEnd, DramahaPhaseRebuy,
	} {
		assert.NotEqual(t, taken, DramahaPhaseDraw)
	}
}

func TestDramaha_ActionConstants(t *testing.T) {
	assert.Equal(t, HoldemActionFold, DramahaActionFold)
	assert.Equal(t, HoldemActionCheck, DramahaActionCheck)
	assert.Equal(t, HoldemActionCall, DramahaActionCall)
	assert.Equal(t, HoldemActionBet, DramahaActionBet)
	assert.Equal(t, HoldemActionRaise, DramahaActionRaise)
	assert.Equal(t, HoldemActionAllIn, DramahaActionAllIn)
}

func TestDramaha_FullGame(t *testing.T) {
	// Play through multiple resets to verify game flow
	o := newTestDramaha()
	for i := 0; i < 3; i++ {
		err := o.Reset()
		assert.NoError(t, err)
		// Verify five hole cards per player
		for j := 0; j < o.GetPlayerCnt(); j++ {
			assert.Equal(t, DramahaHoleCards, o.GetPlayer(j).GetCardsSize())
		}
	}
}

// ---------------------------------------------------------------------------
// Meta-AI integration tests
// ---------------------------------------------------------------------------

func TestDramaha_MetaAI_ProfileSurvivesReset(t *testing.T) {
	o := newTestDramaha()
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)

	_ = o.Reset()
	profile := o.GetHumanProfile()
	assert.NotNil(t, profile, "profile should be created on first Reset with CpuMetaAI=true")
	assert.Equal(t, 0, profile.GamesPlayed)

	_ = o.Reset()
	profile2 := o.GetHumanProfile()
	assert.NotNil(t, profile2)
	assert.Equal(t, 1, profile2.GamesPlayed)
}

func TestDramaha_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	o := newTestDramaha()
	_ = o.Reset()
	assert.Nil(t, o.GetHumanProfile())
}

func TestDramaha_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	o := newTestDramaha()
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)
	_ = o.Reset()
	assert.NotNil(t, o.GetHumanProfile())

	o.ResetProfile()
	assert.Nil(t, o.GetHumanProfile())
}

func TestDramaha_MetaAI_LastHumanPlayMsResetOnReset(t *testing.T) {
	o := setupDramahaForHumanAction(DramahaPhaseFlop)
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)
	o.SetHumanProfile(&BettingHumanProfile{})
	o.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	})
	_ = o.PlayerAction(DramahaActionBet, 20, 800)
	assert.Equal(t, 800, o.GetLastHumanPlayMs(), "lastHumanPlayMs should be set after PlayerAction")

	_ = o.Reset()
	assert.Equal(t, 0, o.GetLastHumanPlayMs(), "lastHumanPlayMs should be reset to 0 on Reset")
}

func TestDramaha_MetaAI_PlayerActionRecordsAction(t *testing.T) {
	t.Run("aggressive action is recorded", func(t *testing.T) {
		o := setupDramahaForHumanAction(DramahaPhaseFlop)
		cfg := o.GetConfig()
		cfg.CpuMetaAI = true
		o.SetConfig(cfg)
		o.SetHumanProfile(&BettingHumanProfile{})
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(DramahaActionBet, 20, 800)
		assert.NoError(t, err)

		profile := o.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 1, profile.HesitationCount)
		total := 0
		for i := 0; i < 3; i++ {
			total += profile.AggressiveByBracket[i].Total
		}
		assert.Equal(t, 1, total)
	})

	t.Run("fold on bet records fold-to-bet", func(t *testing.T) {
		o := setupDramahaForHumanAction(DramahaPhaseFlop)
		cfg := o.GetConfig()
		cfg.CpuMetaAI = true
		o.SetConfig(cfg)
		o.SetHumanProfile(&BettingHumanProfile{})
		o.SetLastBet(40)
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(DramahaActionFold, 0, 0)
		assert.NoError(t, err)

		profile := o.GetHumanProfile()
		assert.Equal(t, 1, profile.FoldToBetCount)
		assert.Equal(t, 1, profile.FoldToBetTotal)
	})

	t.Run("no recording when CpuMetaAI is disabled", func(t *testing.T) {
		o := setupDramahaForHumanAction(DramahaPhaseFlop)
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(DramahaActionBet, 20, 500)
		assert.NoError(t, err)
		assert.Nil(t, o.GetHumanProfile())
	})
}
