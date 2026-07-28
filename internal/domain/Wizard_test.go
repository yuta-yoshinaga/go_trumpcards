//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestWizard() *domain.Wizard {
	players := []*domain.WizardPlayer{
		domain.NewWizardPlayer(true),
		domain.NewWizardPlayer(false),
		domain.NewWizardPlayer(false),
		domain.NewWizardPlayer(false),
	}
	return domain.NewWizard(players, domain.DefaultWizardConfig())
}

func wizardCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func setupWizardBidPhase(o *domain.Wizard, bidPlayerIdx int) {
	o.SetPhase(domain.WizardPhaseBid)
	o.SetBidPlayerIdx(bidPlayerIdx)
}

func setupWizardPlayPhase(o *domain.Wizard, currentIdx, leadIdx, trickNum int) {
	o.SetPhase(domain.WizardPhasePlay)
	o.SetCurrentPlayerIdx(currentIdx)
	o.SetLeadPlayerIdx(leadIdx)
	o.SetTrickNumber(trickNum)
}

// --- Config tests ---

func TestDefaultWizardConfig(t *testing.T) {
	cfg := domain.DefaultWizardConfig()
	assert.Equal(t, domain.WizardCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestWizardConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.WizardConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultWizardConfig(), false},
		{"valid easy", domain.WizardConfig{CpuDifficulty: domain.WizardCpuDifficultyEasy}, false},
		{"valid hard", domain.WizardConfig{CpuDifficulty: domain.WizardCpuDifficultyHard}, false},
		{"invalid difficulty", domain.WizardConfig{CpuDifficulty: 99}, true},
		{"invalid difficulty negative", domain.WizardConfig{CpuDifficulty: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- Player tests ---

func TestWizardPlayer(t *testing.T) {
	p := domain.NewWizardPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, -1, p.GetBid())

	p.SetBid(3)
	assert.Equal(t, 3, p.GetBid())

	p.SetRoundScore(13)
	p.CommitRoundScore()
	assert.Equal(t, 13, p.GetCumulativeScore())

	p.ResetRound()
	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetTrickCount())
}

func TestWizardPlayer_JSON(t *testing.T) {
	p := domain.NewWizardPlayer(true)
	p.SetBid(5)
	p.AddCard(wizardCard(domain.CardDesignSpade, 1))

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	p2 := &domain.WizardPlayer{}
	err = json.Unmarshal(data, p2)
	assert.NoError(t, err)
	assert.Equal(t, 5, p2.GetBid())
	assert.True(t, p2.GetIsHuman())
}

func TestWizardPlayer_UnmarshalJSON_NilGamePlayer(t *testing.T) {
	p := &domain.WizardPlayer{}
	err := json.Unmarshal([]byte(`{}`), p)
	assert.NoError(t, err)
	assert.False(t, p.GetIsHuman())
}

// --- Game tests ---

func TestNewWizard(t *testing.T) {
	o := newTestWizard()
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.Equal(t, -1, o.GetTrumpSuit())
}

func TestNewDefaultWizard(t *testing.T) {
	o := domain.NewDefaultWizard()
	assert.NotNil(t, o)
	assert.Equal(t, domain.WizardPlayerCnt, o.GetPlayerCnt())
	assert.True(t, o.GetPlayer(0).GetIsHuman())
	for i := 1; i < o.GetPlayerCnt(); i++ {
		assert.False(t, o.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.False(t, o.GetGameEndFlag())
}

func TestWizard_Reset(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	assert.Equal(t, domain.WizardPhaseBid, o.GetPhase())
	assert.Equal(t, 1, o.GetRoundNumber())
	assert.Equal(t, 15, o.GetTotalRounds()) // 60 / 4 players
	assert.Equal(t, 1, o.GetHandSize())     // round 1 = 1 card each
	assert.Equal(t, 0, o.GetTrickNumber())
	assert.False(t, o.GetGameEndFlag())
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.Equal(t, 0, o.GetDealerIdx())

	// bidPlayerIdx should be left of dealer (dealer=0, so bidPlayer=1)
	assert.Equal(t, 1, o.GetBidPlayerIdx())

	// each player has 1 card in round 1
	for i := range 4 {
		assert.Equal(t, 1, o.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, o.GetPlayer(i).GetBid())
	}

	// trump card flipped (60 - 4 = 56 remaining → 1 flipped)
	assert.NotNil(t, o.GetTrumpCard())
	assert.GreaterOrEqual(t, o.GetTrumpSuit(), -1)
}

func TestWizard_HandSizeForRound(t *testing.T) {
	// Wizard: round r deals r cards each. Verified indirectly via NextRound.
	o := newTestWizard()
	o.Reset()
	assert.Equal(t, 1, o.GetHandSize())

	o.SetPhase(domain.WizardPhaseRoundEnd)
	o.NextRound()
	assert.Equal(t, 2, o.GetRoundNumber())
	assert.Equal(t, 2, o.GetHandSize())
	for i := range 4 {
		assert.Equal(t, 2, o.GetPlayer(i).GetCardsSize())
	}
}

func TestWizard_FinalRoundNoTrump(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	// Jump to round 15 (final): all 60 cards dealt, no trump card remains.
	o.SetPhase(domain.WizardPhaseRoundEnd)
	o.SetRoundNumber(14)
	o.NextRound() // → round 15

	assert.Equal(t, 15, o.GetRoundNumber())
	assert.Equal(t, 15, o.GetHandSize())
	assert.Nil(t, o.GetTrumpCard())
	assert.Equal(t, -1, o.GetTrumpSuit())
	for i := range 4 {
		assert.Equal(t, 15, o.GetPlayer(i).GetCardsSize())
	}
}

// --- Trump determination tests ---

func TestWizard_ComputeTrumpSuit(t *testing.T) {
	o := newTestWizard()
	o.SetDealerIdx(0)

	t.Run("standard suit", func(t *testing.T) {
		assert.Equal(t, domain.CardDesignHeart, o.ComputeTrumpSuit(wizardCard(domain.CardDesignHeart, 5)))
	})

	t.Run("jester means no trump", func(t *testing.T) {
		assert.Equal(t, -1, o.ComputeTrumpSuit(wizardCard(domain.WizardDesignJester, 1)))
	})

	t.Run("nil means no trump", func(t *testing.T) {
		assert.Equal(t, -1, o.ComputeTrumpSuit(nil))
	})

	t.Run("wizard picks dealer most common suit", func(t *testing.T) {
		dealer := o.GetPlayer(0)
		dealer.Reset()
		dealer.AddCard(wizardCard(domain.CardDesignHeart, 3))
		dealer.AddCard(wizardCard(domain.CardDesignHeart, 7))
		dealer.AddCard(wizardCard(domain.CardDesignHeart, 9))
		dealer.AddCard(wizardCard(domain.CardDesignSpade, 2))
		assert.Equal(t, domain.CardDesignHeart, o.ComputeTrumpSuit(wizardCard(domain.WizardDesignWizard, 1)))
	})

	t.Run("wizard tie picks lowest suit index", func(t *testing.T) {
		dealer := o.GetPlayer(0)
		dealer.Reset()
		dealer.AddCard(wizardCard(domain.CardDesignSpade, 2))
		dealer.AddCard(wizardCard(domain.CardDesignSpade, 5))
		dealer.AddCard(wizardCard(domain.CardDesignHeart, 3))
		dealer.AddCard(wizardCard(domain.CardDesignHeart, 9))
		assert.Equal(t, domain.CardDesignSpade, o.ComputeTrumpSuit(wizardCard(domain.WizardDesignWizard, 1)))
	})

	t.Run("wizard with no suited cards means no trump", func(t *testing.T) {
		dealer := o.GetPlayer(0)
		dealer.Reset()
		dealer.AddCard(wizardCard(domain.WizardDesignWizard, 2))
		dealer.AddCard(wizardCard(domain.WizardDesignJester, 3))
		assert.Equal(t, -1, o.ComputeTrumpSuit(wizardCard(domain.WizardDesignWizard, 1)))
	})
}

// --- Bid tests ---

func TestWizard_PlayerBid_Valid(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetHandSize(10)
	setupWizardBidPhase(o, 0) // human player's turn

	err := o.PlayerBid(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, o.GetPlayer(0).GetBid())
}

func TestWizard_PlayerBid_NoHookRestriction(t *testing.T) {
	// Wizard allows total bids to equal the number of tricks (no hook rule).
	o := newTestWizard()
	o.Reset()
	o.SetHandSize(10)
	o.SetDealerIdx(0)
	setupWizardBidPhase(o, 0)

	o.GetPlayer(1).SetBid(3)
	o.GetPlayer(2).SetBid(2)
	o.GetPlayer(3).SetBid(2)

	// Dealer may bid 3 even though total would equal handSize (10).
	err := o.PlayerBid(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, o.GetPlayer(0).GetBid())
}

func TestWizard_GetRestrictedBid_AlwaysUnrestricted(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	assert.Equal(t, -1, o.GetRestrictedBid())
	o.SetDealerIdx(0)
	setupWizardBidPhase(o, 0)
	assert.Equal(t, -1, o.GetRestrictedBid())
}

func TestWizard_PlayerBid_OutOfRange(t *testing.T) {
	o := newTestWizard()
	o.Reset() // handSize 1
	setupWizardBidPhase(o, 0)

	assert.Error(t, o.PlayerBid(-1))
	assert.Error(t, o.PlayerBid(o.GetHandSize()+1))
}

func TestWizard_PlayerBid_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhasePlay)

	err := o.PlayerBid(1)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestWizard_PlayerBid_NotHumanTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 1) // CPU's turn

	err := o.PlayerBid(1)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestWizard_PlayerBid_GameEnded(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseGameEnd)

	err := o.PlayerBid(1)
	assert.Error(t, err)
}

func TestWizard_CpuBid(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 1) // CPU 1's turn

	o.CpuBid()
	assert.GreaterOrEqual(t, o.GetPlayer(1).GetBid(), 0)
}

func TestWizard_CpuBid_GameEnded(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseGameEnd)
	o.CpuBid() // should not panic
}

func TestWizard_CpuBid_HumanTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 0) // human's turn
	o.CpuBid()                // should do nothing
	assert.Equal(t, -1, o.GetPlayer(0).GetBid())
}

func TestWizard_CpuBid_CountsWizardsAsSureTricks(t *testing.T) {
	o := newTestWizard()
	cfg := domain.DefaultWizardConfig()
	cfg.CpuDifficulty = domain.WizardCpuDifficultyNormal
	o.SetConfig(cfg)
	o.Reset()
	o.SetHandSize(10)

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(wizardCard(domain.WizardDesignWizard, 1))
	p.AddCard(wizardCard(domain.WizardDesignWizard, 2))
	p.AddCard(wizardCard(domain.WizardDesignJester, 1))
	setupWizardBidPhase(o, 1)

	o.CpuBid()
	// 2 wizards → at least 2 sure tricks
	assert.GreaterOrEqual(t, o.GetPlayer(1).GetBid(), 2)
}

func TestWizard_CpuBid_AllDifficulties(t *testing.T) {
	for _, diff := range []domain.WizardCpuDifficulty{
		domain.WizardCpuDifficultyEasy,
		domain.WizardCpuDifficultyNormal,
		domain.WizardCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			o := newTestWizard()
			cfg := domain.DefaultWizardConfig()
			cfg.CpuDifficulty = diff
			o.SetConfig(cfg)
			o.Reset()
			o.SetHandSize(10)
			setupWizardBidPhase(o, 1)

			o.CpuBid()
			bid := o.GetPlayer(1).GetBid()
			assert.GreaterOrEqual(t, bid, 0)
			assert.LessOrEqual(t, bid, o.GetHandSize())
		})
	}
}

// --- Play tests ---

func TestWizard_PlayerPlay_Valid(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))
	setupWizardPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	err := o.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(o.GetCurrentTrick()))
}

func TestWizard_PlayerPlay_InvalidIndex(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardPlayPhase(o, 0, 0, 1)

	assert.Error(t, o.PlayerPlay(-1))
	assert.Error(t, o.PlayerPlay(999))
}

func TestWizard_PlayerPlay_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	err := o.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestWizard_PlayerPlay_GameEnded(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseGameEnd)
	err := o.PlayerPlay(0)
	assert.Error(t, err)
}

func TestWizard_PlayerPlay_FollowSuit(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))

	setupWizardPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 8)},
	})

	// Must follow suit: cannot play spade when holding heart
	assert.Error(t, o.PlayerPlay(1)) // spade
	// Can play heart
	assert.NoError(t, o.PlayerPlay(0)) // heart
}

func TestWizard_PlayerPlay_NoFollowSuitWhenVoid(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))

	setupWizardPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 8)},
	})

	assert.NoError(t, o.PlayerPlay(0))
}

func TestWizard_WizardAndJesterAlwaysPlayable(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))    // lead suit held (index 0)
	p.AddCard(wizardCard(domain.WizardDesignWizard, 1)) // index 1
	p.AddCard(wizardCard(domain.WizardDesignJester, 1)) // index 2
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))   // off-suit (index 3)

	setupWizardPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 8)},
	})

	valid := o.GetValidPlayIndices(0)
	// heart (must follow) + wizard + jester are exempt; spade is not (holds heart)
	assert.Equal(t, []int{0, 1, 2}, valid)
}

func TestWizard_WizardLed_NoFollowObligation(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))

	setupWizardPlayPhase(o, 0, 1, 1)
	// A Wizard was led → no lead suit → anything legal
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.WizardDesignWizard, 1)},
	})

	valid := o.GetValidPlayIndices(0)
	assert.Equal(t, []int{0, 1}, valid)
}

func TestWizard_JesterLed_SuitSetByLaterCard(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))

	setupWizardPlayPhase(o, 0, 1, 1)
	// Jester led then a heart → lead suit is heart, must follow
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.WizardDesignJester, 1)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignHeart, 9)},
	})

	valid := o.GetValidPlayIndices(0)
	assert.Equal(t, []int{0}, valid) // only the heart
}

func TestWizard_CpuPlay(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.SetBid(1)
	setupWizardPlayPhase(o, 1, 1, 1)
	o.SetCurrentTrick(nil)

	o.CpuPlay()
	assert.Equal(t, 1, len(o.GetCurrentTrick()))
}

func TestWizard_CpuPlay_AllDifficulties(t *testing.T) {
	for _, diff := range []domain.WizardCpuDifficulty{
		domain.WizardCpuDifficultyEasy,
		domain.WizardCpuDifficultyNormal,
		domain.WizardCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			o := newTestWizard()
			cfg := domain.DefaultWizardConfig()
			cfg.CpuDifficulty = diff
			o.SetConfig(cfg)
			o.Reset()

			p := o.GetPlayer(1)
			p.Reset()
			p.AddCard(wizardCard(domain.CardDesignHeart, 5))
			p.AddCard(wizardCard(domain.WizardDesignWizard, 1))
			p.AddCard(wizardCard(domain.WizardDesignJester, 1))
			p.SetBid(1)
			setupWizardPlayPhase(o, 1, 1, 1)
			o.SetCurrentTrick(nil)

			o.CpuPlay()
			assert.Equal(t, 1, len(o.GetCurrentTrick()))
		})
	}
}

func TestWizard_CpuPlay_NeedTrickPlaysWizard(t *testing.T) {
	o := newTestWizard()
	cfg := domain.DefaultWizardConfig()
	cfg.CpuDifficulty = domain.WizardCpuDifficultyHard
	o.SetConfig(cfg)
	o.Reset()
	o.SetTrumpSuit(domain.CardDesignSpade)

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 2))
	p.AddCard(wizardCard(domain.WizardDesignWizard, 1))
	p.SetBid(1) // needs a trick

	setupWizardPlayPhase(o, 1, 0, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 13)},
	})

	o.CpuPlay()
	trick := o.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// CPU is void in nothing but should play the wizard to guarantee the win
	assert.Equal(t, domain.WizardDesignWizard, trick[1].Card.GetDesign())
}

func TestWizard_CpuPlay_HumanTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardPlayPhase(o, 0, 0, 1)
	o.CpuPlay() // should do nothing (human's turn)
}

func TestWizard_CpuPlay_GameEnded(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseGameEnd)
	o.CpuPlay() // should not panic
}

func TestWizard_CpuPlay_FollowSuit(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))
	p.SetBid(1)

	setupWizardPlayPhase(o, 1, 0, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 8)},
	})

	o.CpuPlay()
	trick := o.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	assert.Equal(t, domain.CardDesignHeart, trick[1].Card.GetDesign())
}

// --- Trick resolution tests (Wizard rules) ---

func resolveWizardTrick(o *domain.Wizard, trick []*domain.TrickCard) {
	o.SetPhase(domain.WizardPhaseTrickEnd)
	o.SetTrickNumber(1)
	o.SetHandSize(3)
	o.SetCurrentTrick(trick)
	o.ResolveTrick()
}

func TestWizard_TrickWinner_HighestLeadWins(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	// All hearts, no trump — highest lead suit wins
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 5)},
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 3, Card: wizardCard(domain.CardDesignHeart, 3)},
	})
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
	assert.Equal(t, 1, o.GetPlayer(2).GetTrickCount())
}

func TestWizard_TrickWinner_TrumpWins(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetTrumpSuit(domain.CardDesignSpade)
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: wizardCard(domain.CardDesignHeart, 5)},
	})
	// Low trump ♠2 beats high ♥K
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
}

func TestWizard_TrickWinner_WizardBeatsAll(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetTrumpSuit(domain.CardDesignSpade)
	// A Wizard played by player 1 beats even trump aces
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: wizardCard(domain.WizardDesignWizard, 3)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 3, Card: wizardCard(domain.WizardDesignWizard, 4)},
	})
	// First wizard (player 1) wins
	assert.Equal(t, 1, o.GetLeadPlayerIdx())
	assert.Equal(t, 1, o.GetPlayer(1).GetTrickCount())
}

func TestWizard_TrickWinner_JesterNeverWins(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	// No trump; jester led then hearts — highest heart wins, jester loses
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.WizardDesignJester, 1)},
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 4)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignHeart, 11)},
		{PlayerIdx: 3, Card: wizardCard(domain.CardDesignHeart, 2)},
	})
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
}

func TestWizard_TrickWinner_AllJestersFirstWins(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.WizardDesignJester, 1)},
		{PlayerIdx: 1, Card: wizardCard(domain.WizardDesignJester, 2)},
		{PlayerIdx: 2, Card: wizardCard(domain.WizardDesignJester, 3)},
		{PlayerIdx: 3, Card: wizardCard(domain.WizardDesignJester, 4)},
	})
	assert.Equal(t, 0, o.GetLeadPlayerIdx())
	assert.Equal(t, 1, o.GetPlayer(0).GetTrickCount())
}

func TestWizard_TrickWinner_NoTrumpOnlyLeadCounts(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetTrumpSuit(-1) // Reset() flips a random trump; force no-trump so off-suit cards can't win
	// No trump; off-suit high cards cannot win
	resolveWizardTrick(o, []*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 5)},
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 3, Card: wizardCard(domain.CardDesignDiamond, 1)},
	})
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
}

func TestWizard_ResolveTrick_SetsRoundEnd(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseTrickEnd)
	o.SetTrickNumber(3)
	o.SetHandSize(3) // last trick
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 5)},
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 2, Card: wizardCard(domain.CardDesignHeart, 3)},
		{PlayerIdx: 3, Card: wizardCard(domain.CardDesignHeart, 8)},
	})
	o.ResolveTrick()
	assert.Equal(t, domain.WizardPhaseRoundEnd, o.GetPhase())
}

func TestWizard_ResolveTrick_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhasePlay)
	o.ResolveTrick() // should do nothing
}

func TestWizard_ResolveTrick_IncompleteTrick(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseTrickEnd)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: wizardCard(domain.CardDesignHeart, 5)},
	})
	o.ResolveTrick() // should do nothing (not 4 cards)
}

func TestWizard_NextTrick(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseTrickEnd)
	o.SetLeadPlayerIdx(2)
	o.SetTrickNumber(1)

	o.NextTrick()
	assert.Equal(t, domain.WizardPhasePlay, o.GetPhase())
	assert.Equal(t, 2, o.GetCurrentPlayerIdx())
	assert.Equal(t, 2, o.GetTrickNumber())
	assert.Nil(t, o.GetCurrentTrick())
}

func TestWizard_NextTrick_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhasePlay)
	o.NextTrick() // should do nothing
	assert.Equal(t, domain.WizardPhasePlay, o.GetPhase())
}

// --- Scoring tests ---

func TestWizard_ScoreRound_ExactBid(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseRoundEnd)
	o.SetRoundNumber(1)
	o.SetTotalRounds(15)

	o.GetPlayer(0).SetBid(3)
	for range 3 {
		o.GetPlayer(0).AddTrick([]*domain.Card{wizardCard(domain.CardDesignHeart, 5)})
	}
	o.GetPlayer(1).SetBid(0) // 0 tricks = exact

	o.GetPlayer(2).SetBid(2)
	for range 2 {
		o.GetPlayer(2).AddTrick([]*domain.Card{wizardCard(domain.CardDesignHeart, 6)})
	}

	o.GetPlayer(3).SetBid(1)
	o.GetPlayer(3).AddTrick([]*domain.Card{wizardCard(domain.CardDesignHeart, 7)})

	o.ScoreRound()

	assert.Equal(t, 50, o.GetPlayer(0).GetCumulativeScore()) // 20 + 10*3
	assert.Equal(t, 20, o.GetPlayer(1).GetCumulativeScore()) // 20 + 0
	assert.Equal(t, 40, o.GetPlayer(2).GetCumulativeScore()) // 20 + 10*2
	assert.Equal(t, 30, o.GetPlayer(3).GetCumulativeScore()) // 20 + 10*1
	assert.False(t, o.GetGameEndFlag())
}

func TestWizard_ScoreRound_Miss(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseRoundEnd)
	o.SetRoundNumber(1)
	o.SetTotalRounds(15)

	o.GetPlayer(0).SetBid(5)
	for range 2 {
		o.GetPlayer(0).AddTrick([]*domain.Card{wizardCard(domain.CardDesignHeart, 5)})
	}

	o.ScoreRound()
	assert.Equal(t, -30, o.GetPlayer(0).GetCumulativeScore()) // -10*|5-2|
}

func TestWizard_ScoreRound_GameEnd(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseRoundEnd)
	o.SetRoundNumber(15) // last round
	o.SetTotalRounds(15)

	o.GetPlayer(0).SetCumulativeScore(100)
	o.GetPlayer(0).SetBid(0)
	o.GetPlayer(1).SetCumulativeScore(80)
	o.GetPlayer(1).SetBid(0)
	o.GetPlayer(2).SetCumulativeScore(50)
	o.GetPlayer(2).SetBid(0)
	o.GetPlayer(3).SetCumulativeScore(120)
	o.GetPlayer(3).SetBid(0)

	o.ScoreRound()

	assert.True(t, o.GetGameEndFlag())
	assert.Equal(t, domain.WizardPhaseGameEnd, o.GetPhase())
	assert.Equal(t, 3, o.GetWinnerIdx()) // 120 + 20 = 140
}

func TestWizard_ScoreRound_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhasePlay)
	o.ScoreRound() // should do nothing
	assert.Equal(t, domain.WizardPhasePlay, o.GetPhase())
}

// --- NextRound tests ---

func TestWizard_NextRound(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetPhase(domain.WizardPhaseRoundEnd)

	o.NextRound()

	assert.Equal(t, domain.WizardPhaseBid, o.GetPhase())
	assert.Equal(t, 2, o.GetRoundNumber())
	assert.Equal(t, 1, o.GetDealerIdx()) // rotated
	assert.Equal(t, 2, o.GetHandSize())  // round 2 = 2 cards

	for i := range 4 {
		assert.Equal(t, 2, o.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, o.GetPlayer(i).GetBid())
	}
}

func TestWizard_NextRound_WrongPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.NextRound() // should do nothing (phase is Bid)
	assert.Equal(t, 1, o.GetRoundNumber())
}

// --- Hint tests ---

func TestWizard_GetHint_BidPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 0)

	hint := o.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
	assert.Nil(t, hint.CardIndex)
	assert.Equal(t, "strategic_bid", hint.Reason)
}

func TestWizard_GetHint_PlayPhase(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))
	p.SetBid(1)
	setupWizardPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	hint := o.GetHint()
	assert.NotNil(t, hint)
	assert.Nil(t, hint.Bid)
	assert.NotNil(t, hint.CardIndex)
}

func TestWizard_GetHint_PlayPhase_WizardCard(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.WizardDesignWizard, 1))
	p.SetBid(1)
	setupWizardPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	hint := o.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
	assert.Equal(t, "lead_strong", hint.Reason)
}

func TestWizard_GetHint_NotHumanTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 1) // CPU's turn
	assert.Nil(t, o.GetHint())
}

// --- State getters ---

func TestWizard_IsHumanTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetCurrentPlayerIdx(0)
	assert.True(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(1)
	assert.False(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(-1)
	assert.False(t, o.IsHumanTurn())
}

func TestWizard_IsHumanBidTurn(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	o.SetBidPlayerIdx(0)
	assert.True(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(1)
	assert.False(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(-1)
	assert.False(t, o.IsHumanBidTurn())
}

func TestWizard_GetPlayer_OutOfBounds(t *testing.T) {
	o := newTestWizard()
	assert.Nil(t, o.GetPlayer(-1))
	assert.Nil(t, o.GetPlayer(99))
}

func TestWizard_GetPlayerCnt(t *testing.T) {
	o := newTestWizard()
	assert.Equal(t, 4, o.GetPlayerCnt())
}

func TestWizard_ActionLog(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 0)
	_ = o.PlayerBid(1)
	assert.NotEmpty(t, o.GetActionLog())
}

// --- JSON round-trip ---

func TestWizard_JSON_RoundTrip(t *testing.T) {
	o := newTestWizard()
	o.Reset()
	setupWizardBidPhase(o, 0)
	_ = o.PlayerBid(1)

	data, err := json.Marshal(o)
	assert.NoError(t, err)

	o2 := &domain.Wizard{}
	err = json.Unmarshal(data, o2)
	assert.NoError(t, err)

	assert.Equal(t, o.GetPhase(), o2.GetPhase())
	assert.Equal(t, o.GetRoundNumber(), o2.GetRoundNumber())
	assert.Equal(t, o.GetTotalRounds(), o2.GetTotalRounds())
	assert.Equal(t, o.GetHandSize(), o2.GetHandSize())
	assert.Equal(t, o.GetDealerIdx(), o2.GetDealerIdx())
	assert.Equal(t, o.GetTrumpSuit(), o2.GetTrumpSuit())
	assert.Equal(t, o.GetPlayer(0).GetBid(), o2.GetPlayer(0).GetBid())
}

func TestWizard_UnmarshalJSON_OversizedArrays(t *testing.T) {
	o := &domain.Wizard{}
	err := json.Unmarshal([]byte(`{"ps":[`+
		func() string {
			var sb strings.Builder
			for i := range 1001 {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(`{}`)
			}
			return sb.String()
		}()+`]}`), o)
	assert.Error(t, err)
}

func TestWizard_UnmarshalJSON_OversizedDeck(t *testing.T) {
	o := &domain.Wizard{}
	err := json.Unmarshal([]byte(`{"dk":[`+
		func() string {
			var sb strings.Builder
			for i := range domain.WizardDeckSize + 1 {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(`{"d":1,"v":1}`)
			}
			return sb.String()
		}()+`]}`), o)
	assert.Error(t, err)
}

func TestWizard_UnmarshalJSON_NilFields(t *testing.T) {
	o := &domain.Wizard{}
	err := json.Unmarshal([]byte(`{}`), o)
	assert.NoError(t, err)
	assert.Equal(t, 0, o.GetPlayerCnt())
}

func TestWizard_UnmarshalJSON_InvalidIndices(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"currentPlayerIdx out of range", `{"ps":[{},{},{},{}],"ci":99}`},
		{"dealerIdx negative", `{"ps":[{},{},{},{}],"di":-1}`},
		{"bidPlayerIdx out of range", `{"ps":[{},{},{},{}],"bi":99}`},
		{"leadPlayerIdx out of range", `{"ps":[{},{},{},{}],"li":99}`},
		{"trickNumber negative", `{"ps":[{},{},{},{}],"tn":-1}`},
		{"players count not 4", `{"ps":[{},{},{}]}`},
		{"nil player element", `{"ps":[{},null,{},{}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &domain.Wizard{}
			err := json.Unmarshal([]byte(tt.json), o)
			assert.Error(t, err)
		})
	}
}

// --- GetValidPlayIndices ---

func TestWizard_GetValidPlayIndices_FollowSuit(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))
	p.AddCard(wizardCard(domain.CardDesignHeart, 8))

	setupWizardPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: wizardCard(domain.CardDesignHeart, 3)},
	})

	valid := o.GetValidPlayIndices(0)
	assert.Equal(t, []int{0, 2}, valid) // only hearts
}

func TestWizard_GetValidPlayIndices_Lead(t *testing.T) {
	o := newTestWizard()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(wizardCard(domain.CardDesignHeart, 5))
	p.AddCard(wizardCard(domain.CardDesignSpade, 10))

	setupWizardPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	valid := o.GetValidPlayIndices(0)
	assert.Equal(t, []int{0, 1}, valid) // all cards valid when leading
}
