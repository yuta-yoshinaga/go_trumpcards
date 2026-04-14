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
		sd.SetRoundResults([]domain.HoldemResult{
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
		sd.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, WonAmount: 0},
		})
		assert.True(t, sd.IsMuckAvailable())
	})
	t.Run("human won", func(t *testing.T) {
		sd := newTestShortDeck()
		sd.SetPhase(domain.ShortDeckPhaseShowdown)
		sd.SetRoundResults([]domain.HoldemResult{
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

// ---------------------------------------------------------------------------
// Showdown tests
// ---------------------------------------------------------------------------

func TestShortDeck_Showdown(t *testing.T) {
	sd := newTestShortDeck()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).SetChips(1000)
	}
	sd.SetPhase(domain.ShortDeckPhaseRiver)
	sd.SetPot(200)
	sd.SetCurrentTurn(0)
	sd.SetLastBet(0)
	sd.SetMinRaise(10)
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetStartingChips([]int{1000, 1000, 1000, 1000})

	// Player 0: pocket aces (strong)
	sd.GetPlayer(0).Reset()
	sd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	sd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

	// Player 1: weaker hand
	sd.GetPlayer(1).Reset()
	sd.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	sd.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

	// Players 2, 3: folded
	for i := 2; i < 4; i++ {
		sd.GetPlayer(i).SetFolded(true)
		sd.GetPlayer(i).Reset()
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	}

	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	})

	// Human checks → triggers advancePhase → showdown
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	// Handle muck/show
	if sd.IsMuckAvailable() {
		err = sd.Muck()
		assert.NoError(t, err)
	} else if sd.GetPhase() == domain.ShortDeckPhaseShowdown {
		err = sd.ShowHand()
		assert.NoError(t, err)
	}

	assert.True(t, sd.GetGameEndFlag())
	assert.True(t, len(sd.GetRoundResults()) > 0)

	// Verify hand names are populated (covers getHandName)
	for _, r := range sd.GetRoundResults() {
		assert.NotEmpty(t, r.HandName)
	}
}

func TestShortDeck_Showdown_HumanWins(t *testing.T) {
	sd := newTestShortDeck()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).SetChips(1000)
	}
	sd.SetPhase(domain.ShortDeckPhaseRiver)
	sd.SetPot(400)
	sd.SetCurrentTurn(0)
	sd.SetLastBet(0)
	sd.SetMinRaise(10)
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetStartingChips([]int{1000, 1000, 1000, 1000})

	// Player 0 (human): pocket aces → three of a kind with community ace
	sd.GetPlayer(0).Reset()
	sd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	sd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

	// Player 1: weak hand (6, 7 no straight possible)
	sd.GetPlayer(1).Reset()
	sd.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	sd.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 8, false))

	// Players 2, 3: folded
	for i := 2; i < 4; i++ {
		sd.GetPlayer(i).SetFolded(true)
		sd.GetPlayer(i).Reset()
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	}

	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 1, false), // ace → trips for human
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	// Human won → no muck, goes directly to end
	assert.True(t, sd.GetGameEndFlag())
	assert.False(t, sd.IsMuckAvailable())

	// Verify won amount exists
	wonTotal := 0
	for _, r := range sd.GetRoundResults() {
		wonTotal += r.WonAmount
	}
	assert.True(t, wonTotal > 0)
}

// ---------------------------------------------------------------------------
// Phase transition tests (advancePhase + dealRemainingCommunity)
// ---------------------------------------------------------------------------

func TestShortDeck_AdvancePhase_PreFlopToFlop(t *testing.T) {
	sd := newTestShortDeck()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).SetChips(1000)
	}
	sd.SetPhase(domain.ShortDeckPhasePreFlop)
	sd.SetPot(30)
	sd.SetCurrentTurn(0)
	sd.SetLastBet(0)
	sd.SetMinRaise(10)
	sd.SetActedFlags([]bool{false, true, true, true})

	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).Reset()
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	}

	// Human checks → should advance to flop (CPU acts then turn comes back)
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	// Phase should be flop or beyond (CPU might have advanced further)
	assert.True(t, sd.GetPhase() >= domain.ShortDeckPhaseFlop || sd.GetGameEndFlag())
}

func TestShortDeck_AdvancePhase_FlopToTurn(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	// Should advance beyond flop
	assert.True(t, sd.GetPhase() >= domain.ShortDeckPhaseTurn || sd.GetGameEndFlag())
}

func TestShortDeck_AdvancePhase_TurnToRiver(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseTurn)
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	assert.True(t, sd.GetPhase() >= domain.ShortDeckPhaseRiver || sd.GetGameEndFlag())
}

func TestShortDeck_AdvancePhase_RiverToShowdown(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseRiver)
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetStartingChips([]int{1000, 1000, 1000, 1000})
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	assert.True(t, sd.GetPhase() >= domain.ShortDeckPhaseShowdown)
	assert.True(t, len(sd.GetRoundResults()) > 0)
}

func TestShortDeck_DealRemainingCommunity_AllInAtFlop(t *testing.T) {
	sd := newTestShortDeck()
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).SetChips(1000)
	}
	sd.SetPhase(domain.ShortDeckPhaseFlop)
	sd.SetPot(4000)
	sd.SetCurrentTurn(0)
	sd.SetLastBet(0)
	sd.SetMinRaise(10)
	sd.SetStartingChips([]int{1000, 1000, 1000, 1000})

	// All players are all-in except one non-folded
	for i := 0; i < sd.GetPlayerCnt(); i++ {
		sd.GetPlayer(i).Reset()
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		sd.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		if i > 0 {
			sd.GetPlayer(i).SetAllIn(true)
		}
	}

	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	// Human checks → only 1 active (non-all-in) player → dealRemainingCommunity + showdown
	err := sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
	assert.NoError(t, err)

	// Should go to showdown with 5 community cards
	assert.True(t, sd.GetPhase() >= domain.ShortDeckPhaseShowdown)
}

// ---------------------------------------------------------------------------
// GTO CPU player tests
// ---------------------------------------------------------------------------

func TestShortDeck_GTO_FullGame(t *testing.T) {
	// Create a game with GTO-style CPU players to exercise cpuDecidePreFlopGTO and cpuDecidePostFlopGTO
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
	}
	cfg := domain.DefaultShortDeckConfig()
	tc := domain.NewTrumpCardsShortDeck()
	sd := domain.NewShortDeck(tc, players, cfg)

	// Run multiple hands to cover GTO pre-flop and post-flop paths
	for i := 0; i < 10; i++ {
		err := sd.Reset()
		assert.NoError(t, err)

		if sd.GetGameEndFlag() {
			continue
		}

		// Play up to 20 actions per hand to avoid infinite loops
		for attempt := 0; attempt < 20 && !sd.GetGameEndFlag(); attempt++ {
			if !sd.IsHumanTurn() {
				break
			}
			phase := sd.GetPhase()
			if phase < domain.ShortDeckPhasePreFlop || phase > domain.ShortDeckPhaseRiver {
				break
			}
			callAmt := sd.GetLastBet() - sd.GetPlayer(0).GetCurrentBet()
			if callAmt > 0 {
				_ = sd.PlayerAction(domain.ShortDeckActionCall, 0, 0)
			} else {
				_ = sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
			}
		}
		if sd.GetPhase() == domain.ShortDeckPhaseShowdown {
			_ = sd.ShowHand()
		}
	}
}

func TestShortDeck_GTO_PreFlop(t *testing.T) {
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
	}
	cfg := domain.DefaultShortDeckConfig()
	tc := domain.NewTrumpCardsShortDeck()
	sd := domain.NewShortDeck(tc, players, cfg)

	// Run many hands to hit various GTO branches
	for i := 0; i < 20; i++ {
		err := sd.Reset()
		assert.NoError(t, err)

		if sd.GetGameEndFlag() {
			continue
		}
		if sd.IsHumanTurn() && sd.GetPhase() == domain.ShortDeckPhasePreFlop {
			_ = sd.PlayerAction(domain.ShortDeckActionFold, 0, 0)
		}
	}
}

func TestShortDeck_GTO_PostFlop(t *testing.T) {
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
	}
	cfg := domain.DefaultShortDeckConfig()
	tc := domain.NewTrumpCardsShortDeck()
	sd := domain.NewShortDeck(tc, players, cfg)

	// Play enough hands so GTO CPUs reach post-flop
	for i := 0; i < 30; i++ {
		err := sd.Reset()
		assert.NoError(t, err)

		if sd.GetGameEndFlag() {
			continue
		}

		// Keep playing human actions to advance through phases
		for attempt := 0; attempt < 20 && !sd.GetGameEndFlag(); attempt++ {
			if !sd.IsHumanTurn() {
				break
			}
			phase := sd.GetPhase()
			if phase < domain.ShortDeckPhasePreFlop || phase > domain.ShortDeckPhaseRiver {
				break
			}
			if sd.GetLastBet() > sd.GetPlayer(0).GetCurrentBet() {
				_ = sd.PlayerAction(domain.ShortDeckActionCall, 0, 0)
			} else {
				_ = sd.PlayerAction(domain.ShortDeckActionCheck, 0, 0)
			}
		}

		if sd.GetPhase() == domain.ShortDeckPhaseShowdown {
			_ = sd.ShowHand()
		}
	}
}

// ---------------------------------------------------------------------------
// Rebuy / Addon success tests
// ---------------------------------------------------------------------------

func newTestShortDeckWithRebuy() *domain.ShortDeck {
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAG),
	}
	cfg := domain.DefaultShortDeckConfig()
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 3
	cfg.RebuyChips = 1000
	cfg.RebuyPeriodHands = 20
	tc := domain.NewTrumpCardsShortDeck()
	return domain.NewShortDeck(tc, players, cfg)
}

func TestShortDeck_Rebuy_Success(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	// Bust human, trigger rebuy phase
	sd.GetPlayer(0).SetChips(0)
	_ = sd.Reset()
	assert.Equal(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())
	assert.Equal(t, domain.ShortDeckRebuyPhaseRebuy, sd.GetRebuyPhaseType())

	// Execute rebuy
	err := sd.Rebuy()
	assert.NoError(t, err)
	// Human should have chips now
	assert.True(t, sd.GetPlayer(0).GetChips() > 0)
	// Rebuy count incremented
	assert.Equal(t, 1, sd.GetRebuyCounts()[0])
	// Should have continued to deal
	assert.NotEqual(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())
}

func TestShortDeck_SkipRebuy_Success(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	sd.GetPlayer(0).SetChips(0)
	_ = sd.Reset()
	assert.Equal(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())

	err := sd.SkipRebuy()
	assert.NoError(t, err)
	// Human has no chips → game ends
	assert.True(t, sd.GetGameEndFlag())
	assert.Equal(t, domain.ShortDeckPhaseEnd, sd.GetPhase())
}

func TestShortDeck_SkipRebuy_HumanHasChips(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	// CPU busted, not human
	sd.GetPlayer(1).SetChips(0)
	sd.GetPlayer(0).SetChips(1000)
	_ = sd.Reset()

	// If rebuy phase triggered, skip it
	if sd.GetPhase() == domain.ShortDeckPhaseRebuy && sd.GetRebuyPhaseType() == domain.ShortDeckRebuyPhaseRebuy {
		err := sd.SkipRebuy()
		assert.NoError(t, err)
		// Human has chips, game should continue
		assert.False(t, sd.GetGameEndFlag() && sd.GetPhase() == domain.ShortDeckPhaseEnd)
	}
}

func TestShortDeck_Rebuy_ThenAddon(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	cfg := sd.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1 // addon at hand 1
	cfg.AddonChips = 500
	sd.SetConfig(cfg)
	// Bust human
	sd.GetPlayer(0).SetChips(0)
	_ = sd.Reset()
	assert.Equal(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())
	assert.Equal(t, domain.ShortDeckRebuyPhaseRebuy, sd.GetRebuyPhaseType())

	// Do rebuy → should transition to addon phase
	err := sd.Rebuy()
	assert.NoError(t, err)
	assert.Equal(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())
	assert.Equal(t, domain.ShortDeckRebuyPhaseAddon, sd.GetRebuyPhaseType())
}

func TestShortDeck_Addon_Success(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	cfg := sd.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	cfg.AddonChips = 500
	sd.SetConfig(cfg)
	// Bust human to trigger rebuy then addon
	sd.GetPlayer(0).SetChips(0)
	_ = sd.Reset()
	// Rebuy first
	if sd.GetRebuyPhaseType() == domain.ShortDeckRebuyPhaseRebuy {
		_ = sd.Rebuy()
	}
	// Now should be at addon phase
	if sd.GetPhase() == domain.ShortDeckPhaseRebuy && sd.GetRebuyPhaseType() == domain.ShortDeckRebuyPhaseAddon {
		chipsBefore := sd.GetPlayer(0).GetChips()
		err := sd.Addon()
		assert.NoError(t, err)
		assert.Equal(t, chipsBefore+500, sd.GetPlayer(0).GetChips())
		assert.True(t, sd.GetAddonUsed()[0])
	}
}

func TestShortDeck_SkipAddon_Success(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	cfg := sd.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	cfg.AddonChips = 500
	sd.SetConfig(cfg)
	sd.GetPlayer(0).SetChips(0)
	_ = sd.Reset()
	if sd.GetRebuyPhaseType() == domain.ShortDeckRebuyPhaseRebuy {
		_ = sd.Rebuy()
	}
	if sd.GetPhase() == domain.ShortDeckPhaseRebuy && sd.GetRebuyPhaseType() == domain.ShortDeckRebuyPhaseAddon {
		err := sd.SkipAddon()
		assert.NoError(t, err)
		// Should continue to deal
		assert.NotEqual(t, domain.ShortDeckPhaseRebuy, sd.GetPhase())
	}
}

func TestShortDeck_IsRebuyAvailable_Enabled(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	sd.GetPlayer(0).SetChips(0)
	sd.SetHandCount(1)
	assert.True(t, sd.IsRebuyAvailable())
}

func TestShortDeck_IsAddonAvailable_Enabled(t *testing.T) {
	sd := newTestShortDeckWithRebuy()
	cfg := sd.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	cfg.AddonChips = 500
	sd.SetConfig(cfg)
	sd.SetHandCount(1)
	assert.True(t, sd.IsAddonAvailable())
}

// ---------------------------------------------------------------------------
// cpuBetOrAllIn test
// ---------------------------------------------------------------------------

func TestShortDeck_PlayerAction_AllIn(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.GetPlayer(0).SetChips(15) // less than big blind
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	err := sd.PlayerAction(domain.ShortDeckActionAllIn, 0, 0)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// All-fold (resolveLastPlayer) test
// ---------------------------------------------------------------------------

func TestShortDeck_AllFold(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	// Fold players 1-3 manually
	for i := 1; i < 4; i++ {
		sd.GetPlayer(i).SetFolded(true)
	}
	sd.SetActedFlags([]bool{false, true, true, true})
	sd.SetLastBet(50)
	sd.GetPlayer(0).SetCurrentBet(0)

	// Human folds → last remaining player wins
	err := sd.PlayerAction(domain.ShortDeckActionFold, 0, 0)
	assert.NoError(t, err)
	assert.True(t, sd.GetGameEndFlag())
}

// ---------------------------------------------------------------------------
// Tournament blind escalation test
// ---------------------------------------------------------------------------

func TestShortDeck_TournamentBlindEscalation(t *testing.T) {
	sd := newTestShortDeck()
	cfg := sd.GetConfig()
	cfg.TournamentMode = true
	cfg.BlindLevelHands = 2
	cfg.BlindMultiplier = 200 // double
	sd.SetConfig(cfg)

	origSB := cfg.SmallBlind
	origBB := cfg.BigBlind

	// Play hands to trigger blind escalation
	_ = sd.Reset() // hand 1
	_ = sd.Reset() // hand 2 → should escalate

	newCfg := sd.GetConfig()
	assert.True(t, newCfg.SmallBlind >= origSB)
	assert.True(t, newCfg.BigBlind >= origBB)
}

// ---------------------------------------------------------------------------
// Equity and PotOdds
// ---------------------------------------------------------------------------

func TestShortDeck_GetEquity(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	eq := sd.GetEquity()
	assert.NotNil(t, eq)
	assert.True(t, eq.Equity >= 0)
}

func TestShortDeck_GetEquity_WrongPhase(t *testing.T) {
	sd := newTestShortDeck()
	sd.SetPhase(domain.ShortDeckPhaseShowdown)
	assert.Nil(t, sd.GetEquity())
}

func TestShortDeck_GetPotOdds(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})
	sd.SetLastBet(20)
	sd.GetPlayer(0).SetCurrentBet(0)

	odds := sd.GetPotOdds()
	assert.True(t, odds > 0)
}

func TestShortDeck_GetPotOdds_WrongPhase(t *testing.T) {
	sd := newTestShortDeck()
	sd.SetPhase(domain.ShortDeckPhaseShowdown)
	assert.Equal(t, 0.0, sd.GetPotOdds())
}

// ---------------------------------------------------------------------------
// Export/Import profile
// ---------------------------------------------------------------------------

func TestShortDeck_ExportImportProfile(t *testing.T) {
	sd := newTestShortDeck()
	cfg := sd.GetConfig()
	cfg.CpuMetaAI = true
	sd.SetConfig(cfg)
	_ = sd.Reset()

	exported := sd.ExportProfile()
	assert.NotNil(t, exported)

	sd2 := newTestShortDeck()
	assert.Nil(t, sd2.ExportProfile())
}

func TestShortDeck_ImportProfile(t *testing.T) {
	sd := newTestShortDeck()
	err := sd.ImportProfile([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, sd.GetHumanProfile())
}

// ---------------------------------------------------------------------------
// PlayerAction with Call
// ---------------------------------------------------------------------------

func TestShortDeck_PlayerAction_Call(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})
	sd.SetLastBet(20)
	sd.GetPlayer(0).SetCurrentBet(0)

	err := sd.PlayerAction(domain.ShortDeckActionCall, 0, 0)
	assert.NoError(t, err)
}

func TestShortDeck_PlayerAction_Raise(t *testing.T) {
	sd := setupShortDeckForHumanAction(domain.ShortDeckPhaseFlop)
	sd.SetCommunityCards([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})
	sd.SetLastBet(20)
	sd.GetPlayer(0).SetCurrentBet(0)

	err := sd.PlayerAction(domain.ShortDeckActionRaise, 40, 0)
	assert.NoError(t, err)
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
