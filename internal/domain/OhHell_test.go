//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestOhHell() *domain.OhHell {
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	return domain.NewOhHell(domain.NewTrumpCards(0), players, domain.DefaultOhHellConfig())
}

func setupOhHellBidPhase(o *domain.OhHell, bidPlayerIdx int) {
	o.SetPhase(domain.OhHellPhaseBid)
	o.SetBidPlayerIdx(bidPlayerIdx)
}

func setupOhHellPlayPhase(o *domain.OhHell, currentIdx, leadIdx, trickNum int) {
	o.SetPhase(domain.OhHellPhasePlay)
	o.SetCurrentPlayerIdx(currentIdx)
	o.SetLeadPlayerIdx(leadIdx)
	o.SetTrickNumber(trickNum)
}

// --- Config tests ---

func TestDefaultOhHellConfig(t *testing.T) {
	cfg := domain.DefaultOhHellConfig()
	assert.Equal(t, domain.OhHellCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 10, cfg.MaxHandSize)
	assert.Equal(t, domain.OhHellScoringStandard, cfg.ScoringVariant)
	assert.Equal(t, domain.OhHellRoundDownAndUp, cfg.RoundDirection)
}

func TestOhHellConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.OhHellConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultOhHellConfig(), false},
		{"invalid difficulty", domain.OhHellConfig{CpuDifficulty: 99, MaxHandSize: 10}, true},
		{"invalid max hand size 0", domain.OhHellConfig{MaxHandSize: 0}, true},
		{"invalid max hand size 14", domain.OhHellConfig{MaxHandSize: 14}, true},
		{"valid max hand size 1", domain.OhHellConfig{MaxHandSize: 1}, false},
		{"valid max hand size 13", domain.OhHellConfig{MaxHandSize: 13}, false},
		{"invalid scoring variant", domain.OhHellConfig{MaxHandSize: 10, ScoringVariant: 99}, true},
		{"invalid round direction", domain.OhHellConfig{MaxHandSize: 10, RoundDirection: 99}, true},
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

func TestOhHellPlayer(t *testing.T) {
	p := domain.NewOhHellPlayer(true)
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

func TestOhHellPlayer_JSON(t *testing.T) {
	p := domain.NewOhHellPlayer(true)
	p.SetBid(5)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	p2 := &domain.OhHellPlayer{}
	err = json.Unmarshal(data, p2)
	assert.NoError(t, err)
	assert.Equal(t, 5, p2.GetBid())
	assert.True(t, p2.GetIsHuman())
}

func TestOhHellPlayer_UnmarshalJSON_NilGamePlayer(t *testing.T) {
	p := &domain.OhHellPlayer{}
	err := json.Unmarshal([]byte(`{}`), p)
	assert.NoError(t, err)
	assert.False(t, p.GetIsHuman())
}

// --- Game tests ---

func TestNewOhHell(t *testing.T) {
	o := newTestOhHell()
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.Equal(t, -1, o.GetTrumpSuit())
}

func TestNewDefaultOhHell(t *testing.T) {
	o := domain.NewDefaultOhHell()
	assert.NotNil(t, o)
	assert.Equal(t, domain.OhHellPlayerCnt, o.GetPlayerCnt())
	assert.True(t, o.GetPlayer(0).GetIsHuman())
	for i := 1; i < o.GetPlayerCnt(); i++ {
		assert.False(t, o.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.False(t, o.GetGameEndFlag())
}

func TestOhHell_Reset(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	assert.Equal(t, domain.OhHellPhaseBid, o.GetPhase())
	assert.Equal(t, 1, o.GetRoundNumber())
	assert.Equal(t, 19, o.GetTotalRounds()) // DownAndUp: 10*2-1
	assert.Equal(t, 10, o.GetHandSize())    // round 1 = max hand size
	assert.Equal(t, 0, o.GetTrickNumber())
	assert.False(t, o.GetGameEndFlag())
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.Equal(t, 0, o.GetDealerIdx())

	// bidPlayerIdx should be left of dealer (dealer=0, so bidPlayer=1)
	assert.Equal(t, 1, o.GetBidPlayerIdx())

	// 全プレイヤーに10枚ずつ配られている
	for i := range 4 {
		assert.Equal(t, 10, o.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, o.GetPlayer(i).GetBid())
	}

	// 切り札カードが公開されている (52 - 40 = 12枚残り → 1枚めくれる)
	assert.NotNil(t, o.GetTrumpCard())
	assert.GreaterOrEqual(t, o.GetTrumpSuit(), 0)
}

func TestOhHell_Reset_DownOnly(t *testing.T) {
	o := newTestOhHell()
	cfg := domain.DefaultOhHellConfig()
	cfg.RoundDirection = domain.OhHellRoundDownOnly
	o.SetConfig(cfg)
	o.Reset()

	assert.Equal(t, 10, o.GetTotalRounds()) // DownOnly: 10
}

func TestOhHell_HandSizeForRound(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	// DownAndUp with MaxHandSize=10: round 1=10, round 10=1, round 11=2, round 19=10
	tests := []struct {
		round    int
		expected int
	}{
		{1, 10}, {2, 9}, {5, 6}, {10, 1}, {11, 2}, {15, 6}, {19, 10},
	}
	for _, tt := range tests {
		t.Run("round_"+string(rune('0'+tt.round/10))+string(rune('0'+tt.round%10)), func(t *testing.T) {
			o.SetRoundNumber(tt.round)
			// We can't call handSizeForRound directly, but we can test via NextRound behavior
		})
	}
}

func TestOhHell_NoTrumpWhenAllDealt(t *testing.T) {
	// 4 players × 13 cards = 52 cards = entire deck
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	cfg := domain.DefaultOhHellConfig()
	cfg.MaxHandSize = 13
	o := domain.NewOhHell(domain.NewTrumpCards(0), players, cfg)
	o.Reset()

	// When handSize=13, all 52 cards are dealt, no trump card remains
	assert.Equal(t, 13, o.GetHandSize())
	assert.Nil(t, o.GetTrumpCard())
	assert.Equal(t, -1, o.GetTrumpSuit())
}

// --- Bid tests ---

func TestOhHell_PlayerBid_Valid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0) // human player's turn

	err := o.PlayerBid(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, o.GetPlayer(0).GetBid())
}

func TestOhHell_PlayerBid_OutOfRange(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0)

	assert.Error(t, o.PlayerBid(-1))
	assert.Error(t, o.PlayerBid(o.GetHandSize()+1))
}

func TestOhHell_PlayerBid_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhasePlay)

	err := o.PlayerBid(3)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestOhHell_PlayerBid_NotHumanTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 1) // CPU's turn

	err := o.PlayerBid(3)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestOhHell_PlayerBid_GameEnded(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseGameEnd)

	err := o.PlayerBid(3)
	assert.Error(t, err) // ErrWrongPhase (gameEndFlag not set via SetPhase)
}

func TestOhHell_DealerRestriction(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	// Set dealer=0 (human), handSize=10
	o.SetDealerIdx(0)
	setupOhHellBidPhase(o, 0)

	// Set other players' bids to total 7
	o.GetPlayer(1).SetBid(3)
	o.GetPlayer(2).SetBid(2)
	o.GetPlayer(3).SetBid(2)

	// Dealer cannot bid 3 (would make total = 10 = handSize)
	err := o.PlayerBid(3)
	assert.Error(t, err)

	// Dealer can bid 2 or 4
	err = o.PlayerBid(2)
	assert.NoError(t, err)
}

func TestOhHell_GetRestrictedBid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetDealerIdx(0)
	setupOhHellBidPhase(o, 0)

	o.GetPlayer(1).SetBid(3)
	o.GetPlayer(2).SetBid(4)
	o.GetPlayer(3).SetBid(1)

	// Total bids so far = 8, handSize = 10, restricted = 10 - 8 = 2
	assert.Equal(t, 2, o.GetRestrictedBid())
}

func TestOhHell_GetRestrictedBid_NotDealer(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0)
	// bidPlayer=0, dealer=0 by default after Reset
	// But if dealer != bidPlayer, no restriction
	o.SetDealerIdx(1)

	assert.Equal(t, -1, o.GetRestrictedBid())
}

func TestOhHell_CpuBid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 1) // CPU 1's turn

	o.CpuBid()
	assert.GreaterOrEqual(t, o.GetPlayer(1).GetBid(), 0)
}

func TestOhHell_CpuBid_GameEnded(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseGameEnd)

	o.CpuBid() // should not panic
}

func TestOhHell_CpuBid_HumanTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0) // human's turn

	o.CpuBid() // should do nothing
	assert.Equal(t, -1, o.GetPlayer(0).GetBid())
}

func TestOhHell_CpuBid_DealerRestriction(t *testing.T) {
	// CPU as dealer should respect restriction
	o := newTestOhHell()
	o.Reset()
	o.SetDealerIdx(1)
	setupOhHellBidPhase(o, 1)

	// Set other bids so only one value is restricted
	o.GetPlayer(0).SetBid(5)
	o.GetPlayer(2).SetBid(3)
	o.GetPlayer(3).SetBid(1)
	// total = 9, handSize=10, restricted = 1

	// Run many times to verify CPU never bids the restricted value
	for range 100 {
		o.GetPlayer(1).SetBid(-1)
		setupOhHellBidPhase(o, 1)
		o.CpuBid()
		assert.NotEqual(t, 1, o.GetPlayer(1).GetBid())
	}
}

// --- Play tests ---

func TestOhHell_PlayerPlay_Valid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	// Set up play phase manually
	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	setupOhHellPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	err := o.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(o.GetCurrentTrick()))
}

func TestOhHell_PlayerPlay_InvalidIndex(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellPlayPhase(o, 0, 0, 1)

	assert.Error(t, o.PlayerPlay(-1))
	assert.Error(t, o.PlayerPlay(999))
}

func TestOhHell_PlayerPlay_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	err := o.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestOhHell_PlayerPlay_GameEnded(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseGameEnd)

	err := o.PlayerPlay(0)
	assert.Error(t, err) // ErrWrongPhase (gameEndFlag not set via SetPhase)
}

func TestOhHell_PlayerPlay_FollowSuit(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	setupOhHellPlayPhase(o, 0, 1, 1)
	// Lead card is a heart
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	})

	// Must follow suit: cannot play spade when holding heart
	err := o.PlayerPlay(1) // spade
	assert.Error(t, err)

	// Can play heart
	err = o.PlayerPlay(0) // heart
	assert.NoError(t, err)
}

func TestOhHell_PlayerPlay_NoFollowSuitWhenVoid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	setupOhHellPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	})

	// No hearts in hand, can play anything
	err := o.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestOhHell_CpuPlay(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.SetBid(1)
	setupOhHellPlayPhase(o, 1, 1, 1)
	o.SetCurrentTrick(nil)

	o.CpuPlay()
	assert.Equal(t, 1, len(o.GetCurrentTrick()))
}

func TestOhHell_CpuPlay_AllDifficulties(t *testing.T) {
	for _, diff := range []domain.OhHellCpuDifficulty{
		domain.OhHellCpuDifficultyEasy,
		domain.OhHellCpuDifficultyNormal,
		domain.OhHellCpuDifficultyHard,
	} {
		t.Run("difficulty_"+string(rune('0'+int(diff))), func(t *testing.T) {
			o := newTestOhHell()
			cfg := domain.DefaultOhHellConfig()
			cfg.CpuDifficulty = diff
			o.SetConfig(cfg)
			o.Reset()

			p := o.GetPlayer(1)
			p.Reset()
			p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
			p.SetBid(1)
			setupOhHellPlayPhase(o, 1, 1, 1)
			o.SetCurrentTrick(nil)

			o.CpuPlay()
			assert.Equal(t, 1, len(o.GetCurrentTrick()))
		})
	}
}

// --- Trick resolution tests ---

func TestOhHell_ResolveTrick_HighestLeadWins(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetTrickNumber(1)
	o.SetHandSize(3)
	// All hearts — trump irrelevant, highest lead suit wins
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
	})

	o.ResolveTrick()

	// Player 2 wins (highest heart K=13)
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
	assert.Equal(t, 1, o.GetPlayer(2).GetTrickCount())
}

func TestOhHell_ResolveTrick_TrumpWins(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	// Explicitly set trump to spade for deterministic testing
	o.SetTrumpSuit(domain.CardDesignSpade)

	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetTrickNumber(1)
	o.SetHandSize(3)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 2, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})

	o.ResolveTrick()
	// Player 2 wins with trump (even low ♠2 beats high ♥K)
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
	assert.Equal(t, 1, o.GetPlayer(2).GetTrickCount())
}

func TestOhHell_ResolveTrick_NoTrump(t *testing.T) {
	// When there's no trump, only lead suit matters
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	cfg := domain.DefaultOhHellConfig()
	cfg.MaxHandSize = 13
	o := domain.NewOhHell(domain.NewTrumpCards(0), players, cfg)
	o.Reset()
	// trumpSuit = -1 when all cards dealt

	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetTrickNumber(1)
	o.SetHandSize(13)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignDiamond, 1, false)},
	})

	o.ResolveTrick()
	// Player 2 wins (highest of lead suit = heart)
	assert.Equal(t, 2, o.GetLeadPlayerIdx())
}

func TestOhHell_ResolveTrick_SetsRoundEnd(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetTrickNumber(10) // last trick (handSize=10)
	o.SetHandSize(10)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	})

	o.ResolveTrick()
	assert.Equal(t, domain.OhHellPhaseRoundEnd, o.GetPhase())
}

func TestOhHell_NextTrick(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetLeadPlayerIdx(2)
	o.SetTrickNumber(1)

	o.NextTrick()
	assert.Equal(t, domain.OhHellPhasePlay, o.GetPhase())
	assert.Equal(t, 2, o.GetCurrentPlayerIdx())
	assert.Equal(t, 2, o.GetTrickNumber())
	assert.Nil(t, o.GetCurrentTrick())
}

func TestOhHell_NextTrick_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhasePlay)

	o.NextTrick() // should do nothing
	assert.Equal(t, domain.OhHellPhasePlay, o.GetPhase())
}

// --- Scoring tests ---

func TestOhHell_ScoreRound_ExactBid(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseRoundEnd)
	o.SetRoundNumber(1)
	o.SetTotalRounds(19)

	o.GetPlayer(0).SetBid(3)
	// Add 3 tricks
	for range 3 {
		o.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	}

	o.GetPlayer(1).SetBid(0)
	// 0 tricks = exact

	o.GetPlayer(2).SetBid(2)
	for range 2 {
		o.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
	}

	o.GetPlayer(3).SetBid(5)
	for range 5 {
		o.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	}

	o.ScoreRound()

	assert.Equal(t, 13, o.GetPlayer(0).GetCumulativeScore()) // 10 + 3
	assert.Equal(t, 10, o.GetPlayer(1).GetCumulativeScore()) // 10 + 0
	assert.Equal(t, 12, o.GetPlayer(2).GetCumulativeScore()) // 10 + 2
	assert.Equal(t, 15, o.GetPlayer(3).GetCumulativeScore()) // 10 + 5
	assert.False(t, o.GetGameEndFlag())                      // not last round
}

func TestOhHell_ScoreRound_Miss_Standard(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseRoundEnd)
	o.SetRoundNumber(1)
	o.SetTotalRounds(19)

	o.GetPlayer(0).SetBid(5)
	// only 2 tricks taken
	for range 2 {
		o.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	}

	o.ScoreRound()
	assert.Equal(t, 0, o.GetPlayer(0).GetCumulativeScore()) // standard: 0 for miss
}

func TestOhHell_ScoreRound_Miss_Penalty(t *testing.T) {
	o := newTestOhHell()
	cfg := domain.DefaultOhHellConfig()
	cfg.ScoringVariant = domain.OhHellScoringPenalty
	o.SetConfig(cfg)
	o.Reset()
	o.SetPhase(domain.OhHellPhaseRoundEnd)
	o.SetRoundNumber(1)
	o.SetTotalRounds(19)

	o.GetPlayer(0).SetBid(5)
	for range 2 {
		o.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	}

	o.ScoreRound()
	assert.Equal(t, -3, o.GetPlayer(0).GetCumulativeScore()) // -|5-2| = -3
}

func TestOhHell_ScoreRound_GameEnd(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseRoundEnd)
	o.SetRoundNumber(19) // last round
	o.SetTotalRounds(19)

	// Set cumulative scores and bids
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
	assert.Equal(t, domain.OhHellPhaseGameEnd, o.GetPhase())
	assert.Equal(t, 3, o.GetWinnerIdx()) // player 3 has 120 + 10 = 130
}

func TestOhHell_ScoreRound_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhasePlay)

	o.ScoreRound() // should do nothing
	assert.Equal(t, domain.OhHellPhasePlay, o.GetPhase())
}

// --- NextRound tests ---

func TestOhHell_NextRound(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseRoundEnd)

	o.NextRound()

	assert.Equal(t, domain.OhHellPhaseBid, o.GetPhase())
	assert.Equal(t, 2, o.GetRoundNumber())
	assert.Equal(t, 1, o.GetDealerIdx()) // rotated
	assert.Equal(t, 9, o.GetHandSize())  // round 2 = 9

	for i := range 4 {
		assert.Equal(t, 9, o.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, o.GetPlayer(i).GetBid())
	}
}

func TestOhHell_NextRound_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	o.NextRound() // should do nothing (phase is Bid)
	assert.Equal(t, 1, o.GetRoundNumber())
}

// --- Hint tests ---

func TestOhHell_GetHint_BidPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0)

	hint := o.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
	assert.Nil(t, hint.CardIndex)
	assert.Equal(t, "strategic_bid", hint.Reason)
}

func TestOhHell_GetHint_PlayPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.SetBid(1)
	setupOhHellPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	hint := o.GetHint()
	assert.NotNil(t, hint)
	assert.Nil(t, hint.Bid)
	assert.NotNil(t, hint.CardIndex)
}

func TestOhHell_GetHint_NotHumanTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 1) // CPU's turn

	hint := o.GetHint()
	assert.Nil(t, hint)
}

// --- State getters ---

func TestOhHell_IsHumanTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetCurrentPlayerIdx(0)
	assert.True(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(1)
	assert.False(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(-1)
	assert.False(t, o.IsHumanTurn())
}

func TestOhHell_IsHumanBidTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetBidPlayerIdx(0)
	assert.True(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(1)
	assert.False(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(-1)
	assert.False(t, o.IsHumanBidTurn())
}

func TestOhHell_GetPlayer_OutOfBounds(t *testing.T) {
	o := newTestOhHell()
	assert.Nil(t, o.GetPlayer(-1))
	assert.Nil(t, o.GetPlayer(99))
}

func TestOhHell_GetPlayerCnt(t *testing.T) {
	o := newTestOhHell()
	assert.Equal(t, 4, o.GetPlayerCnt())
}

func TestOhHell_ActionLog(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0)

	_ = o.PlayerBid(2)
	assert.NotEmpty(t, o.GetActionLog())
}

// --- JSON round-trip ---

func TestOhHell_JSON_RoundTrip(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellBidPhase(o, 0)
	_ = o.PlayerBid(3)

	data, err := json.Marshal(o)
	assert.NoError(t, err)

	o2 := &domain.OhHell{}
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

func TestOhHell_UnmarshalJSON_OversizedArrays(t *testing.T) {
	// Create a JSON with >1000 players to trigger size guard
	o := &domain.OhHell{}
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

func TestOhHell_UnmarshalJSON_NilFields(t *testing.T) {
	o := &domain.OhHell{}
	err := json.Unmarshal([]byte(`{}`), o)
	assert.NoError(t, err)
	assert.Equal(t, 0, o.GetPlayerCnt())
}

func TestOhHell_UnmarshalJSON_InvalidIndices(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"currentPlayerIdx out of range", `{"ps":[{},{},{},{}],"ci":99}`},
		{"dealerIdx negative", `{"ps":[{},{},{},{}],"di":-1}`},
		{"bidPlayerIdx out of range", `{"ps":[{},{},{},{}],"bi":99}`},
		{"leadPlayerIdx out of range", `{"ps":[{},{},{},{}],"li":99}`},
		{"trickNumber negative", `{"ps":[{},{},{},{}],"tn":-1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &domain.OhHell{}
			err := json.Unmarshal([]byte(tt.json), o)
			assert.Error(t, err)
		})
	}
}

// --- CPU bidding difficulties ---

func TestOhHell_CpuBid_Easy(t *testing.T) {
	o := newTestOhHell()
	cfg := domain.DefaultOhHellConfig()
	cfg.CpuDifficulty = domain.OhHellCpuDifficultyEasy
	o.SetConfig(cfg)
	o.Reset()
	setupOhHellBidPhase(o, 1)

	o.CpuBid()
	bid := o.GetPlayer(1).GetBid()
	assert.GreaterOrEqual(t, bid, 0)
	assert.LessOrEqual(t, bid, o.GetHandSize())
}

func TestOhHell_CpuBid_Normal(t *testing.T) {
	o := newTestOhHell()
	cfg := domain.DefaultOhHellConfig()
	cfg.CpuDifficulty = domain.OhHellCpuDifficultyNormal
	o.SetConfig(cfg)
	o.Reset()
	setupOhHellBidPhase(o, 1)

	o.CpuBid()
	bid := o.GetPlayer(1).GetBid()
	assert.GreaterOrEqual(t, bid, 0)
	assert.LessOrEqual(t, bid, o.GetHandSize())
}

func TestOhHell_CpuBid_Hard(t *testing.T) {
	o := newTestOhHell()
	cfg := domain.DefaultOhHellConfig()
	cfg.CpuDifficulty = domain.OhHellCpuDifficultyHard
	o.SetConfig(cfg)
	o.Reset()
	setupOhHellBidPhase(o, 1)

	o.CpuBid()
	bid := o.GetPlayer(1).GetBid()
	assert.GreaterOrEqual(t, bid, 0)
	assert.LessOrEqual(t, bid, o.GetHandSize())
}

// --- GetValidPlayIndices ---

func TestOhHell_GetValidPlayIndices(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	setupOhHellPlayPhase(o, 0, 1, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
	})

	valid := o.GetValidPlayIndices(0)
	// Should only include hearts (indices 0 and 2)
	assert.Equal(t, []int{0, 2}, valid)
}

func TestOhHell_GetValidPlayIndices_Lead(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	setupOhHellPlayPhase(o, 0, 0, 1)
	o.SetCurrentTrick(nil)

	valid := o.GetValidPlayIndices(0)
	// All cards valid when leading
	assert.Equal(t, []int{0, 1}, valid)
}

// --- ResolveTrick edge cases ---

func TestOhHell_ResolveTrick_WrongPhase(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhasePlay)

	o.ResolveTrick() // should do nothing
}

func TestOhHell_ResolveTrick_IncompleteTrick(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseTrickEnd)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})

	o.ResolveTrick() // should do nothing (not 4 cards)
}

// --- CPU play with follow scenarios ---

func TestOhHell_CpuPlay_FollowSuit(t *testing.T) {
	o := newTestOhHell()
	o.Reset()

	p := o.GetPlayer(1)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.SetBid(1)

	setupOhHellPlayPhase(o, 1, 0, 1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	})

	o.CpuPlay()
	trick := o.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// CPU must follow heart suit
	assert.Equal(t, domain.CardDesignHeart, trick[1].Card.GetDesign())
}

func TestOhHell_CpuPlay_HumanTurn(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	setupOhHellPlayPhase(o, 0, 0, 1)

	o.CpuPlay() // should do nothing (human's turn)
}

func TestOhHell_CpuPlay_GameEnded(t *testing.T) {
	o := newTestOhHell()
	o.Reset()
	o.SetPhase(domain.OhHellPhaseGameEnd)

	o.CpuPlay() // should not panic
}

// --- Scoring-variant-aware bidding (#5514) ---

// setupOhHellBidHand は指定席に指定の手札だけを持たせ、入札フェーズに置く。
// ディーラーを席1にするのは、ディーラー制限 (合計ビッド == handSize) が
// 見積もりそのものを ±1 ずらして測定を汚すため。
func setupOhHellBidHand(o *domain.OhHell, seat int, cards []*domain.Card) {
	o.Reset()
	o.SetDealerIdx(1)
	setupOhHellBidPhase(o, seat)
	for i := range 4 {
		o.GetPlayer(i).Reset()
		o.GetPlayer(i).SetBid(-1)
	}
	p := o.GetPlayer(seat)
	for _, c := range cards {
		p.AddCard(c)
	}
	o.SetHandSize(len(cards))
	o.SetTrumpSuit(domain.CardDesignSpade)
}

// ohHellSpeculativeHand は「確実な札」と「あてずっぽうの札」が両方入った手札。
// 切り札3枚 (長さボーナスの条件) + 非切り札のQ (ハードCPUがコイントスで
// 数える札) + 非切り札のK (確実な札)。
func ohHellSpeculativeHand() []*domain.Card {
	return []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
	}
}

// ohHellHintBid は同じ手札でヒントのビッドを何度も引き、出た値の集合を返す。
// cpuBidHard はQをコイントスで数えるので、1回引いただけでは
// 「方式で変わった」のか「たまたま裏が出た」のか区別できない。
func ohHellHintBids(t *testing.T, variant domain.OhHellScoringVariant) map[int]bool {
	t.Helper()
	seen := map[int]bool{}
	for range 200 {
		o := newTestOhHell()
		cfg := domain.DefaultOhHellConfig()
		cfg.ScoringVariant = variant
		o.SetConfig(cfg)
		setupOhHellBidHand(o, 0, ohHellSpeculativeHand())
		hint := o.GetHint()
		if assert.NotNil(t, hint) && assert.NotNil(t, hint.Bid) {
			seen[*hint.Bid] = true
		}
	}
	return seen
}

// ペナルティ方式では外したぶんだけ減点されるので、ヒントは
// あてずっぽうの札を数えない ― しかも毎回同じ数を返す。
func TestOhHell_GetHint_PenaltyBidIsConservativeAndStable(t *testing.T) {
	standard := ohHellHintBids(t, domain.OhHellScoringStandard)
	penalty := ohHellHintBids(t, domain.OhHellScoringPenalty)

	// 標準方式: Qのコイントスで2通り出る (200回引いて両方見えないほうがおかしい)
	assert.Len(t, standard, 2, "standard hint should still be a coin toss: %v", standard)
	// ペナルティ方式: 確実な札だけなので毎回同じ
	assert.Len(t, penalty, 1, "penalty hint should be deterministic: %v", penalty)

	for p := range penalty {
		for s := range standard {
			assert.Less(t, p, s, "penalty bid %d must be lower than standard bid %d", p, s)
		}
	}
}

// ノーマルCPUも同じ手札で慎重になる。ハードが50%扱いする非切り札のQを
// ノーマルは確定扱いしていたので、そこが方式で分かれる。
func TestOhHell_CpuBidNormal_PenaltyDropsTheCoinTossCard(t *testing.T) {
	bidFor := func(variant domain.OhHellScoringVariant) int {
		o := newTestOhHell()
		cfg := domain.DefaultOhHellConfig()
		cfg.ScoringVariant = variant
		cfg.CpuDifficulty = domain.OhHellCpuDifficultyNormal
		o.SetConfig(cfg)
		setupOhHellBidHand(o, 2, ohHellSpeculativeHand())
		o.CpuBid()
		return o.GetPlayer(2).GetBid()
	}

	standard := bidFor(domain.OhHellScoringStandard)
	penalty := bidFor(domain.OhHellScoringPenalty)
	assert.Equal(t, 4, standard, "trump J/10 + off-suit Q/K")
	assert.Equal(t, 3, penalty, "the off-suit Q is a guess, so penalty scoring drops it")
}

// 負のコントロール: 確実な札しか無い手札なら方式で差が出ない。
// 「ペナルティだと必ず下がる」ではなく「あてずっぽうを数えない」が仕様。
func TestOhHell_CpuBidNormal_CertainCardsAreCountedInBothVariants(t *testing.T) {
	certain := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
	}
	bidFor := func(variant domain.OhHellScoringVariant) int {
		o := newTestOhHell()
		cfg := domain.DefaultOhHellConfig()
		cfg.ScoringVariant = variant
		cfg.CpuDifficulty = domain.OhHellCpuDifficultyNormal
		o.SetConfig(cfg)
		setupOhHellBidHand(o, 2, certain)
		o.CpuBid()
		return o.GetPlayer(2).GetBid()
	}
	assert.Equal(t, 2, bidFor(domain.OhHellScoringStandard))
	assert.Equal(t, 2, bidFor(domain.OhHellScoringPenalty))
}
