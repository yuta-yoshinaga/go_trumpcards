package domain_test

import (
	"testing"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func newTestShortDeck() *domain.ShortDeck {
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAG),
	}
	cfg := domain.DefaultShortDeckConfig()
	tc := domain.NewTrumpCardsShortDeck()
	return domain.NewShortDeck(tc, players, cfg)
}

func setupShortDeckForHumanAction(phase int) *domain.ShortDeck {
	sd := newTestShortDeck()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).SetChips(1000)
	}
	sd.SetPhase(phase)
	sd.SetCurrentTurn(0)
	sd.SetLastBet(0)
	sd.SetMinRaise(10)
	sd.SetPot(30)
	// Give players 2 cards each (ShortDeck deals 2 hole cards)
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		p := sd.GetPlayer(i)
		p.Reset()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	}
	return sd
}

func TestNewShortDeck(t *testing.T) {
	sd := newTestShortDeck()
	assert.Equal(t, domain.ShortDeckPhaseInit, sd.GetPhase())
	assert.Equal(t, 4, sd.GetPlayerCnt())
	assert.NotNil(t, sd.GetCommunityCards())
	assert.Equal(t, 0, sd.GetPot())
	assert.False(t, sd.GetGameEndFlag())
}

func TestShortDeck_Reset(t *testing.T) {
	sd := newTestShortDeck()
	_ = sd.Reset()

	// After reset, each player should have 2 cards (ShortDeck)
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		p := sd.GetPlayer(i)
		assert.Equal(t, 2, p.GetCardsSize(), "player %d should have 2 cards", i)
		assert.True(t, p.GetChips() > 0 || p.GetAllIn())
	}
	assert.True(t, sd.GetPot() > 0)
}

func TestShortDeck_Reset_Deals2Cards(t *testing.T) {
	sd := newTestShortDeck()
	_ = sd.Reset()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		assert.Equal(t, 2, sd.GetPlayer(i).GetCardsSize())
	}
}

func TestShortDeck_Resize(t *testing.T) {
	sd := newTestShortDeck()
	assert.Equal(t, 4, sd.GetPlayerCnt())

	newPlayers := make([]*domain.ShortDeckPlayer, 6)
	newPlayers[0] = domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP)
	}
	sd.Resize(newPlayers)
	assert.Equal(t, 6, sd.GetPlayerCnt())
	assert.Equal(t, 0, sd.GetHandCount())
	assert.Equal(t, 6, len(sd.GetActedFlags()))

	err := sd.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 6, sd.GetPlayerCnt())
	for i := 0; i < 6; i++ {
		assert.Equal(t, 2, sd.GetPlayer(i).GetCardsSize())
	}
}

func TestShortDeck_PlayerAction_Fold(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})
	sd.SetLastBet(50)
	sd.GetPlayer(0).SetCurrentBet(0)

	err := sd.PlayerAction(domain.ShortDeckActionFold, 0, 0)
	assert.NoError(t, err)
}

func TestShortDeck_PlayerAction_Check(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestShortDeck_PlayerAction_Bet(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionBet, 20, 0)
	assert.NoError(t, err)
}

func TestShortDeck_PlayerAction_GameEnded(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetGameEndFlag(true)
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestShortDeck_PlayerAction_WrongPhase(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseShowdown)
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestShortDeck_PlayerAction_NotHumanTurn(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})
	sd.SetCurrentTurn(1) // CPU
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.Error(t, err)
}

func TestShortDeck_Getters(t *testing.T) {
	sd := newTestShortDeck()
	_ = sd.Reset()

	assert.NotNil(t, sd.GetCommunityCards())
	assert.NotNil(t, sd.GetSidePots())
	assert.Equal(t, 0, sd.GetDealerIdx())
	assert.True(t, sd.GetLastBet() >= 0)
	assert.NotNil(t, sd.GetRoundResults())
	assert.NotNil(t, sd.GetCpuActions())
	assert.Nil(t, sd.GetLastCpuError())
	assert.Equal(t, 4, sd.GetConfig().TableSize)
	assert.NotNil(t, sd.GetActedFlags())
	assert.Equal(t, 1, sd.GetHandCount())
	assert.NotNil(t, sd.GetActionLog())
	assert.NotNil(t, sd.GetRebuyCounts())
	assert.NotNil(t, sd.GetAddonUsed())
	assert.Equal(t, domain.ShortDeckRebuyPhaseNone, sd.GetRebuyPhaseType())
}

func TestShortDeck_GetPlayer_InvalidIndex(t *testing.T) {
	sd := newTestShortDeck()
	assert.Nil(t, sd.GetPlayer(-1))
	assert.Nil(t, sd.GetPlayer(100))
}

func TestShortDeck_IsHumanTurn(t *testing.T) {
	sd := newTestShortDeck()
	_ = sd.Reset()
	sd.SetCurrentTurn(0)
	assert.True(t, sd.IsHumanTurn())
	sd.SetCurrentTurn(1)
	assert.False(t, sd.IsHumanTurn())
}

func TestShortDeck_SetConfig(t *testing.T) {
	sd := newTestShortDeck()
	cfg := domain.DefaultShortDeckConfig()
	cfg.BigBlind = 20
	sd.SetConfig(cfg)
	assert.Equal(t, 20, sd.GetConfig().BigBlind)
}

func TestShortDeck_Muck(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.Muck())
	})
	t.Run("showdown phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseShowdown)
		sd.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, WonAmount: 0},
		})
		err := sd.Muck()
		assert.NoError(t, err)
		assert.Equal(t, domain.ShortDeckPhaseEnd, sd.GetPhase())
	})
}

func TestShortDeck_ShowHand(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.ShowHand())
	})
	t.Run("showdown phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseShowdown)
		err := sd.ShowHand()
		assert.NoError(t, err)
		assert.Equal(t, domain.ShortDeckPhaseEnd, sd.GetPhase())
	})
}

func TestShortDeck_IsMuckAvailable(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.False(t, sd.IsMuckAvailable())
	})
	t.Run("human lost", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseShowdown)
		sd.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, WonAmount: 0},
		})
		assert.True(t, sd.IsMuckAvailable())
	})
	t.Run("human won", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseShowdown)
		sd.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, WonAmount: 100},
		})
		assert.False(t, sd.IsMuckAvailable())
	})
}

func TestShortDeck_Rebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.Rebuy())
	})
	t.Run("wrong rebuy type", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseRebuy)
		sd.SetRebuyPhaseType(domain.ShortDeckRebuyPhaseAddon)
		assert.Error(t, sd.Rebuy())
	})
}

func TestShortDeck_SkipRebuy(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.SkipRebuy())
	})
}

func TestShortDeck_Addon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.Addon())
	})
	t.Run("wrong addon type", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseRebuy)
		sd.SetRebuyPhaseType(domain.ShortDeckRebuyPhaseRebuy)
		assert.Error(t, sd.Addon())
	})
}

func TestShortDeck_SkipAddon(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseFlop)
		assert.Error(t, sd.SkipAddon())
	})
}

func TestShortDeck_IsRebuyAvailable(t *testing.T) {
	t.Run("rebuy disabled", func(t *testing.T) {
		sd := newTestShortDeck()
		assert.False(t, sd.IsRebuyAvailable())
	})
}

func TestShortDeck_IsAddonAvailable(t *testing.T) {
	t.Run("addon disabled", func(t *testing.T) {
		sd := newTestShortDeck()
		assert.False(t, sd.IsAddonAvailable())
	})
}

func TestShortDeck_PhaseConstants(t *testing.T) {
	assert.Equal(t, 0, domain.ShortDeckPhaseInit)
	assert.Equal(t, 1, domain.ShortDeckPhasePreFlop)
	assert.Equal(t, 2, domain.ShortDeckPhaseFlop)
	assert.Equal(t, 3, domain.ShortDeckPhaseTurn)
	assert.Equal(t, 4, domain.ShortDeckPhaseRiver)
	assert.Equal(t, 5, domain.ShortDeckPhaseShowdown)
	assert.Equal(t, 6, domain.ShortDeckPhaseEnd)
	assert.Equal(t, 7, domain.ShortDeckPhaseRebuy)
}

func TestShortDeck_ActionConstants(t *testing.T) {
	assert.Equal(t, domain.HoldemActionFold, domain.ShortDeckActionFold)
	assert.Equal(t, domain.HoldemActionCheck, domain.ShortDeckActionCheck)
	assert.Equal(t, domain.HoldemActionCall, domain.ShortDeckActionCall)
	assert.Equal(t, domain.HoldemActionBet, domain.ShortDeckActionBet)
	assert.Equal(t, domain.HoldemActionRaise, domain.ShortDeckActionRaise)
	assert.Equal(t, domain.HoldemActionAllIn, domain.ShortDeckActionAllIn)
}

func TestShortDeck_FullGame(t *testing.T) {
	// Play through multiple resets to verify game flow
	sd := newTestShortDeck()
	for i := 0; i < 3; i++ {
		err := sd.Reset()
		assert.NoError(t, err)
		// Verify 2 cards per player (ShortDeck deals 2 hole cards)
		for j := 0; j < sd.GetPlayerCnt(); j++ {
			assert.Equal(t, 2, sd.GetPlayer(j).GetCardsSize())
		}
	}
}

func TestShortDeck_36CardDeck(t *testing.T) {
	// Verify the ShortDeck deck is 36 cards (values 6-A across 4 suits)
	tc := domain.NewTrumpCardsShortDeck()
	assert.Equal(t, 36, tc.GetTotalCount())
}

func TestShortDeck_FlushBeatsFullHouseInShowdown(t *testing.T) {
	// In ShortDeck, Flush (rank 6) beats FullHouse (rank 5)
	assert.Greater(t, domain.ShortDeckHandFlush, domain.ShortDeckHandFullHouse)
}

// ---------------------------------------------------------------------------
// Meta-AI integration tests
// ---------------------------------------------------------------------------

func TestShortDeck_MetaAI_ProfileSurvivesReset(t *testing.T) {
	sd := newTestShortDeck()
	cfg := sd.GetConfig()
	cfg.CpuMetaAI = true
	sd.SetConfig(cfg)

	_ = sd.Reset()
	profile := sd.GetHumanProfile()
	assert.NotNil(t, profile, "profile should be created on first Reset with CpuMetaAI=true")
	assert.Equal(t, 0, profile.GamesPlayed)

	_ = sd.Reset()
	profile2 := sd.GetHumanProfile()
	assert.NotNil(t, profile2)
	assert.Equal(t, 1, profile2.GamesPlayed)
}

func TestShortDeck_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	sd := newTestShortDeck()
	_ = sd.Reset()
	assert.Nil(t, sd.GetHumanProfile())
}

func TestShortDeck_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	sd := newTestShortDeck()
	cfg := sd.GetConfig()
	cfg.CpuMetaAI = true
	sd.SetConfig(cfg)
	_ = sd.Reset()
	assert.NotNil(t, sd.GetHumanProfile())

	sd.ResetProfile()
	assert.Nil(t, sd.GetHumanProfile())
}

func TestShortDeck_MetaAI_LastHumanPlayMsResetOnReset(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	cfg := sd.GetConfig()
	cfg.CpuMetaAI = true
	sd.SetConfig(cfg)
	sd.SetHumanProfile(&domain.BettingHumanProfile{})
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	})
	_ = sd.PlayerAction(domain.ShortDeckActionBet, 20, 800)
	assert.Equal(t, 800, sd.GetLastHumanPlayMs(), "lastHumanPlayMs should be set after PlayerAction")

	_ = sd.Reset()
	assert.Equal(t, 0, sd.GetLastHumanPlayMs(), "lastHumanPlayMs should be reset to 0 on Reset")
}

func TestShortDeck_MetaAI_PlayerActionRecordsAction(t *testing.T) {
	t.Run("aggressive action is recorded", func(t *testing.T) {
		sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
		cfg := sd.GetConfig()
		cfg.CpuMetaAI = true
		sd.SetConfig(cfg)
		sd.SetHumanProfile(&domain.BettingHumanProfile{})
		sd.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		err := sd.PlayerAction(domain.ShortDeckActionBet, 20, 800)
		assert.NoError(t, err)

		profile := sd.GetHumanProfile()
		assert.NotNil(t, profile)
		assert.Equal(t, 1, profile.HesitationCount)
		total := 0
		for i := 0; i < 3; i++ {
			total += profile.AggressiveByBracket[i].Total
		}
		assert.Equal(t, 1, total)
	})

	t.Run("fold on bet records fold-to-bet", func(t *testing.T) {
		sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
		cfg := sd.GetConfig()
		cfg.CpuMetaAI = true
		sd.SetConfig(cfg)
		sd.SetHumanProfile(&domain.BettingHumanProfile{})
		sd.SetLastBet(40)
		sd.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		err := sd.PlayerAction(domain.ShortDeckActionFold, 0, 0)
		assert.NoError(t, err)

		profile := sd.GetHumanProfile()
		assert.Equal(t, 1, profile.FoldToBetCount)
		assert.Equal(t, 1, profile.FoldToBetTotal)
	})

	t.Run("no recording when CpuMetaAI is disabled", func(t *testing.T) {
		sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
		sd.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		err := sd.PlayerAction(domain.ShortDeckActionBet, 20, 500)
		assert.NoError(t, err)
		assert.Nil(t, sd.GetHumanProfile())
	})
}
