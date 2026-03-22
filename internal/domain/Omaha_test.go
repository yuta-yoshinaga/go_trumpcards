package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestOmaha() *Omaha {
	players := []*OmahaPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleLAP),
		NewOmahaPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultOmahaConfig()
	tc := NewTrumpCards(0)
	return NewOmaha(tc, players, cfg)
}

func setupOmahaForHumanAction(phase int) *Omaha {
	o := newTestOmaha()
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
	// Give players 4 cards each
	for _, p := range o.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 13, false))
		p.AddCard(NewCard(CardDesignClover, 12, false))
		p.AddCard(NewCard(CardDesignDiamond, 11, false))
	}
	return o
}

func TestNewOmaha(t *testing.T) {
	o := newTestOmaha()
	assert.Equal(t, OmahaPhaseInit, o.GetPhase())
	assert.Equal(t, 4, o.GetPlayerCnt())
	assert.NotNil(t, o.GetCommunityCards())
	assert.Equal(t, 0, o.GetPot())
	assert.False(t, o.GetGameEndFlag())
}

func TestOmaha_Reset(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()

	// After reset, each player should have 4 cards (Omaha)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		p := o.GetPlayer(i)
		assert.Equal(t, 4, p.GetCardsSize(), "player %d should have 4 cards", i)
		assert.True(t, p.GetChips() > 0 || p.GetAllIn())
	}
	assert.True(t, o.GetPot() > 0)
}

func TestOmaha_Reset_Deals4Cards(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, 4, o.GetPlayer(i).GetCardsSize())
	}
}

func TestOmaha_Resize(t *testing.T) {
	o := newTestOmaha()
	assert.Equal(t, 4, o.GetPlayerCnt())

	newPlayers := make([]*OmahaPlayer, 6)
	newPlayers[0] = NewOmahaPlayer(true, HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = NewOmahaPlayer(false, HoldemStyleLAP)
	}
	o.Resize(newPlayers)
	assert.Equal(t, 6, o.GetPlayerCnt())
	assert.Equal(t, 0, o.GetHandCount())
	assert.Equal(t, 6, len(o.GetActedFlags()))

	err := o.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 6, o.GetPlayerCnt())
	for i := 0; i < 6; i++ {
		assert.Equal(t, 4, o.GetPlayer(i).GetCardsSize())
	}
}

func TestOmaha_PlayerAction_Fold(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.lastBet = 50
	o.players[0].SetCurrentBet(0)

	err := o.PlayerAction(OmahaActionFold, 0, 0)
	assert.NoError(t, err)
}

func TestOmaha_PlayerAction_Check(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestOmaha_PlayerAction_Bet(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}

	err := o.PlayerAction(OmahaActionBet, 20, 0)
	assert.NoError(t, err)
}

func TestOmaha_PlayerAction_GameEnded(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.gameEndFlag = true
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_PlayerAction_WrongPhase(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseShowdown)
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_PlayerAction_NotHumanTurn(t *testing.T) {
	o := setupOmahaForHumanAction(OmahaPhaseFlop)
	o.communityCards = []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	o.currentTurn = 1 // CPU
	err := o.PlayerAction(OmahaActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestOmaha_Getters(t *testing.T) {
	o := newTestOmaha()
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
	assert.Equal(t, OmahaRebuyPhaseNone, o.GetRebuyPhaseType())
}

func TestOmaha_GetPlayer_InvalidIndex(t *testing.T) {
	o := newTestOmaha()
	assert.Nil(t, o.GetPlayer(-1))
	assert.Nil(t, o.GetPlayer(100))
}

func TestOmaha_IsHumanTurn(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	o.currentTurn = 0
	assert.True(t, o.IsHumanTurn())
	o.currentTurn = 1
	assert.False(t, o.IsHumanTurn())
}

func TestOmaha_SetConfig(t *testing.T) {
	o := newTestOmaha()
	cfg := DefaultOmahaConfig()
	cfg.BigBlind = 20
	o.SetConfig(cfg)
	assert.Equal(t, 20, o.GetConfig().BigBlind)
}

func TestOmaha_Muck(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Muck())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []OmahaResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		err := o.Muck()
		assert.NoError(t, err)
		assert.Equal(t, OmahaPhaseEnd, o.phase)
	})
}

func TestOmaha_ShowHand(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.ShowHand())
	})
	t.Run("showdown phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		err := o.ShowHand()
		assert.NoError(t, err)
		assert.Equal(t, OmahaPhaseEnd, o.phase)
	})
}

func TestOmaha_IsMuckAvailable(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.False(t, o.IsMuckAvailable())
	})
	t.Run("human lost", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []OmahaResult{
			{PlayerIdx: 0, WonAmount: 0},
		}
		assert.True(t, o.IsMuckAvailable())
	})
	t.Run("human won", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseShowdown
		o.roundResults = []OmahaResult{
			{PlayerIdx: 0, WonAmount: 100},
		}
		assert.False(t, o.IsMuckAvailable())
	})
}

func TestOmaha_Rebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Rebuy())
	})
	t.Run("wrong rebuy type", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseRebuy
		o.rebuyPhaseType = OmahaRebuyPhaseAddon
		assert.Error(t, o.Rebuy())
	})
}

func TestOmaha_SkipRebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.SkipRebuy())
	})
}

func TestOmaha_Addon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.Addon())
	})
	t.Run("wrong addon type", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseRebuy
		o.rebuyPhaseType = OmahaRebuyPhaseRebuy
		assert.Error(t, o.Addon())
	})
}

func TestOmaha_SkipAddon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		o := newTestOmaha()
		o.phase = OmahaPhaseFlop
		assert.Error(t, o.SkipAddon())
	})
}

func TestOmaha_IsRebuyAvailable(t *testing.T) {
	t.Run("rebuy disabled", func(t *testing.T) {
		o := newTestOmaha()
		assert.False(t, o.IsRebuyAvailable())
	})
}

func TestOmaha_IsAddonAvailable(t *testing.T) {
	t.Run("addon disabled", func(t *testing.T) {
		o := newTestOmaha()
		assert.False(t, o.IsAddonAvailable())
	})
}

func TestOmaha_PhaseConstants(t *testing.T) {
	assert.Equal(t, 0, OmahaPhaseInit)
	assert.Equal(t, 1, OmahaPhasePreFlop)
	assert.Equal(t, 2, OmahaPhaseFlop)
	assert.Equal(t, 3, OmahaPhaseTurn)
	assert.Equal(t, 4, OmahaPhaseRiver)
	assert.Equal(t, 5, OmahaPhaseShowdown)
	assert.Equal(t, 6, OmahaPhaseEnd)
	assert.Equal(t, 7, OmahaPhaseRebuy)
}

func TestOmaha_ActionConstants(t *testing.T) {
	assert.Equal(t, HoldemActionFold, OmahaActionFold)
	assert.Equal(t, HoldemActionCheck, OmahaActionCheck)
	assert.Equal(t, HoldemActionCall, OmahaActionCall)
	assert.Equal(t, HoldemActionBet, OmahaActionBet)
	assert.Equal(t, HoldemActionRaise, OmahaActionRaise)
	assert.Equal(t, HoldemActionAllIn, OmahaActionAllIn)
}

func TestOmaha_FullGame(t *testing.T) {
	// Play through multiple resets to verify game flow
	o := newTestOmaha()
	for i := 0; i < 3; i++ {
		err := o.Reset()
		assert.NoError(t, err)
		// Verify 4 cards per player
		for j := 0; j < o.GetPlayerCnt(); j++ {
			assert.Equal(t, 4, o.GetPlayer(j).GetCardsSize())
		}
	}
}

// ---------------------------------------------------------------------------
// Meta-AI integration tests
// ---------------------------------------------------------------------------

func TestOmaha_MetaAI_ProfileSurvivesReset(t *testing.T) {
	o := newTestOmaha()
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

func TestOmaha_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	o := newTestOmaha()
	_ = o.Reset()
	assert.Nil(t, o.GetHumanProfile())
}

func TestOmaha_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	o := newTestOmaha()
	cfg := o.GetConfig()
	cfg.CpuMetaAI = true
	o.SetConfig(cfg)
	_ = o.Reset()
	assert.NotNil(t, o.GetHumanProfile())

	o.ResetProfile()
	assert.Nil(t, o.GetHumanProfile())
}

func TestOmaha_MetaAI_PlayerActionRecordsAction(t *testing.T) {
	t.Run("aggressive action is recorded", func(t *testing.T) {
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
		cfg := o.GetConfig()
		cfg.CpuMetaAI = true
		o.SetConfig(cfg)
		o.SetHumanProfile(&BettingHumanProfile{})
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(OmahaActionBet, 20, 800)
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
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
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

		err := o.PlayerAction(OmahaActionFold, 0, 0)
		assert.NoError(t, err)

		profile := o.GetHumanProfile()
		assert.Equal(t, 1, profile.FoldToBetCount)
		assert.Equal(t, 1, profile.FoldToBetTotal)
	})

	t.Run("no recording when CpuMetaAI is disabled", func(t *testing.T) {
		o := setupOmahaForHumanAction(OmahaPhaseFlop)
		o.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignDiamond, 9, false),
		})

		err := o.PlayerAction(OmahaActionBet, 20, 500)
		assert.NoError(t, err)
		assert.Nil(t, o.GetHumanProfile())
	})
}
