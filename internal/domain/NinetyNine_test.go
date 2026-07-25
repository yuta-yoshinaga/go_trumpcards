//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestNinetyNine() *domain.NinetyNine {
	players := []*domain.NinetyNinePlayer{
		domain.NewNinetyNinePlayer(true),
		domain.NewNinetyNinePlayer(false),
		domain.NewNinetyNinePlayer(false),
	}
	return domain.NewNinetyNine(domain.NewTrumpCardsNinetyNine(), players, domain.DefaultNinetyNineConfig())
}

// --- Deck ---

func TestNewTrumpCardsNinetyNine_36Distinct(t *testing.T) {
	deck := domain.NewTrumpCardsNinetyNine()
	assert.Equal(t, 36, deck.GetTotalCount())

	seen := make(map[string]bool)
	count := 0
	for {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		key := strings.Join([]string{string(rune(c.GetDesign())), string(rune(c.GetValue()))}, "-")
		assert.False(t, seen[key], "duplicate card detected")
		seen[key] = true
		count++
	}
	assert.Equal(t, 36, count)
	assert.Len(t, seen, 36)
}

func TestNinetyNine_DealLeavesZero(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	total := 0
	for i := 0; i < o.GetPlayerCnt(); i++ {
		// after bury (cpus bid in Reset, human still 12), count cards + buried
		p := o.GetPlayer(i)
		total += p.GetCardsSize() + len(p.GetBuried())
	}
	// 3 players * 12 dealt = 36
	assert.Equal(t, 36, total)
}

// --- Config ---

func TestDefaultNinetyNineConfig(t *testing.T) {
	cfg := domain.DefaultNinetyNineConfig()
	assert.Equal(t, domain.NinetyNineCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 100, cfg.TargetScore)
}

func TestNinetyNineConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.NinetyNineConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultNinetyNineConfig(), false},
		{"invalid difficulty", domain.NinetyNineConfig{CpuDifficulty: 99, TargetScore: 100}, true},
		{"target too low", domain.NinetyNineConfig{TargetScore: 9}, true},
		{"target too high", domain.NinetyNineConfig{TargetScore: 1001}, true},
		{"valid target 10", domain.NinetyNineConfig{TargetScore: 10}, false},
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

// --- Player ---

func TestNinetyNinePlayer(t *testing.T) {
	p := domain.NewNinetyNinePlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, -1, p.GetBid())

	p.SetBid(3)
	assert.Equal(t, 3, p.GetBid())

	p.SetBuried([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	assert.Len(t, p.GetBuried(), 1)

	p.SetRoundScore(13)
	p.CommitRoundScore()
	assert.Equal(t, 13, p.GetCumulativeScore())

	p.ResetRound()
	assert.Equal(t, -1, p.GetBid())
	assert.Nil(t, p.GetBuried())
	assert.Equal(t, 0, p.GetRoundScore())
}

func TestNinetyNinePlayer_JSON(t *testing.T) {
	p := domain.NewNinetyNinePlayer(true)
	p.SetBid(5)
	p.SetBuried([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	p2 := &domain.NinetyNinePlayer{}
	require.NoError(t, json.Unmarshal(data, p2))
	assert.Equal(t, 5, p2.GetBid())
	assert.True(t, p2.GetIsHuman())
	assert.Len(t, p2.GetBuried(), 1)
}

func TestNinetyNinePlayer_UnmarshalJSON_NilGamePlayer(t *testing.T) {
	p := &domain.NinetyNinePlayer{}
	require.NoError(t, json.Unmarshal([]byte(`{}`), p))
	assert.False(t, p.GetIsHuman())
}

// --- Construction / Reset ---

func TestNewNinetyNine(t *testing.T) {
	o := newTestNinetyNine()
	assert.Equal(t, -1, o.GetWinnerIdx())
	assert.Equal(t, domain.NinetyNinePlayerCnt, o.GetPlayerCnt())
}

func TestNewDefaultNinetyNine(t *testing.T) {
	o := domain.NewDefaultNinetyNine()
	assert.Equal(t, 3, o.GetPlayerCnt())
	assert.True(t, o.GetPlayer(0).GetIsHuman())
	assert.False(t, o.GetPlayer(1).GetIsHuman())
}

func TestNinetyNine_Reset(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	assert.Equal(t, domain.NinetyNinePhaseBid, o.GetPhase())
	assert.Equal(t, 1, o.GetDealNumber())
	assert.Equal(t, 9, o.GetHandSize())
	assert.False(t, o.GetGameEndFlag())
	// trump suit for deal 1 = spade
	assert.Equal(t, domain.CardDesignSpade, o.GetTrumpSuit())
	// human dealt 12 cards
	assert.Equal(t, 12, o.GetPlayer(1).GetCardsSize())
}

// --- Bury → bid (suit sum) ---

func setupBidWithKnownHand(o *domain.NinetyNine, designs []int) {
	// Replace human (idx0) hand deterministically
	p := o.GetPlayer(0)
	p.Reset()
	for _, d := range designs {
		p.AddCard(domain.NewCard(d, 7, false))
	}
	o.SetPhase(domain.NinetyNinePhaseBid)
	o.SetBidPlayerIdx(0)
}

func TestNinetyNine_PlayerBid_SuitSum(t *testing.T) {
	cases := []struct {
		name    string
		designs []int // first 3 are buried
		bury    []int
		wantBid int
	}{
		{"all diamonds = 0", []int{domain.CardDesignDiamond, domain.CardDesignDiamond, domain.CardDesignDiamond, domain.CardDesignSpade}, []int{0, 1, 2}, 0},
		{"all spades = 3", []int{domain.CardDesignSpade, domain.CardDesignSpade, domain.CardDesignSpade, domain.CardDesignDiamond}, []int{0, 1, 2}, 3},
		{"all hearts = 6", []int{domain.CardDesignHeart, domain.CardDesignHeart, domain.CardDesignHeart, domain.CardDesignDiamond}, []int{0, 1, 2}, 6},
		{"all clubs = 9", []int{domain.CardDesignClover, domain.CardDesignClover, domain.CardDesignClover, domain.CardDesignDiamond}, []int{0, 1, 2}, 9},
		{"mixed D+S+H = 3", []int{domain.CardDesignDiamond, domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignClover}, []int{0, 1, 2}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := newTestNinetyNine()
			setupBidWithKnownHand(o, tc.designs)
			require.NoError(t, o.PlayerBid(tc.bury))
			assert.Equal(t, tc.wantBid, o.GetPlayer(0).GetBid())
			assert.Len(t, o.GetPlayer(0).GetBuried(), 3)
			assert.Equal(t, len(tc.designs)-3, o.GetPlayer(0).GetCardsSize())
		})
	}
}

func TestNinetyNine_PlayerBid_Errors(t *testing.T) {
	o := newTestNinetyNine()
	setupBidWithKnownHand(o, []int{domain.CardDesignSpade, domain.CardDesignSpade, domain.CardDesignSpade, domain.CardDesignSpade})

	// wrong count
	assert.Error(t, o.PlayerBid([]int{0, 1}))
	// duplicate
	assert.Error(t, o.PlayerBid([]int{0, 0, 1}))
	// out of range
	assert.Error(t, o.PlayerBid([]int{0, 1, 99}))

	o.SetPhase(domain.NinetyNinePhasePlay)
	assert.Error(t, o.PlayerBid([]int{0, 1, 2}))

	o.SetPhase(domain.NinetyNinePhaseBid)
	o.SetBidPlayerIdx(1) // not human
	assert.Error(t, o.PlayerBid([]int{0, 1, 2}))
}

func TestNinetyNine_CpuBid(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	// move bid pointer to a CPU
	o.SetBidPlayerIdx(1)
	o.CpuBid()
	assert.GreaterOrEqual(t, o.GetPlayer(1).GetBid(), 0)
	assert.LessOrEqual(t, o.GetPlayer(1).GetBid(), 9)
	assert.Equal(t, 9, o.GetPlayer(1).GetCardsSize())
}

func TestNinetyNine_BidPhaseAdvancesToPlay(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	// All three bid; loop CPUs then human last
	for i := 0; i < 3; i++ {
		if o.GetBidPlayerIdx() == 0 {
			require.NoError(t, o.PlayerBid([]int{0, 1, 2}))
		} else {
			o.CpuBid()
		}
	}
	assert.Equal(t, domain.NinetyNinePhasePlay, o.GetPhase())
	assert.Equal(t, 1, o.GetTrickNumber())
}

// --- Play / must-follow ---

func setupPlay(o *domain.NinetyNine) {
	o.SetPhase(domain.NinetyNinePhasePlay)
	o.SetTrumpSuit(domain.CardDesignSpade)
	for i := 0; i < 3; i++ {
		o.GetPlayer(i).Reset()
	}
}

func TestNinetyNine_PlayerPlay_MustFollow(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetCurrentPlayerIdx(0)
	o.SetLeadPlayerIdx(0)
	// human has hearts and a club; lead heart trick already in progress
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
	})
	// playing club (idx1) while holding heart must fail
	assert.Error(t, o.PlayerPlay(1))
	// playing heart ok
	assert.NoError(t, o.PlayerPlay(0))
}

func TestNinetyNine_PlayerPlay_LeadAnything(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetCurrentPlayerIdx(0)
	o.SetCurrentTrick(nil)
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	assert.NoError(t, o.PlayerPlay(0))
}

func TestNinetyNine_CpuPlay(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	// drive to play phase
	for i := 0; i < 3; i++ {
		if o.GetBidPlayerIdx() == 0 {
			require.NoError(t, o.PlayerBid([]int{0, 1, 2}))
		} else {
			o.CpuBid()
		}
	}
	require.Equal(t, domain.NinetyNinePhasePlay, o.GetPhase())
	// run until human turn or trick end
	for o.GetPhase() == domain.NinetyNinePhasePlay && !o.GetPlayer(o.GetCurrentPlayerIdx()).GetIsHuman() {
		before := o.GetCurrentPlayerIdx()
		o.CpuPlay()
		assert.NotEqual(t, before, o.GetCurrentPlayerIdx())
	}
}

// --- Trick resolution ---

func TestNinetyNine_ResolveTrick_HighestLeadWins(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetTrumpSuit(domain.CardDesignClover) // no trump in trick
	o.SetTrickNumber(1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})
	o.SetPhase(domain.NinetyNinePhaseTrickEnd)
	o.ResolveTrick()
	assert.Equal(t, 1, o.GetPlayer(1).GetTrickCount())
	assert.Equal(t, 1, o.GetLeadPlayerIdx())
}

func TestNinetyNine_ResolveTrick_AceHighWins(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetTrumpSuit(domain.CardDesignClover)
	o.SetTrickNumber(1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // Ace
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 12, false)},
	})
	o.SetPhase(domain.NinetyNinePhaseTrickEnd)
	o.ResolveTrick()
	assert.Equal(t, 1, o.GetPlayer(1).GetTrickCount())
}

func TestNinetyNine_ResolveTrick_TrumpWins(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetTrumpSuit(domain.CardDesignSpade)
	o.SetTrickNumber(1)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // Ace heart lead
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 6, false)}, // low trump
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
	})
	o.SetPhase(domain.NinetyNinePhaseTrickEnd)
	o.ResolveTrick()
	assert.Equal(t, 1, o.GetPlayer(1).GetTrickCount())
}

func TestNinetyNine_ResolveTrick_LastTrickSetsRoundEnd(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetTrickNumber(domain.NinetyNineTricksPerDeal)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})
	o.SetPhase(domain.NinetyNinePhaseTrickEnd)
	o.ResolveTrick()
	assert.Equal(t, domain.NinetyNinePhaseRoundEnd, o.GetPhase())
}

func TestNinetyNine_NextTrick(t *testing.T) {
	o := newTestNinetyNine()
	o.SetPhase(domain.NinetyNinePhaseTrickEnd)
	o.SetLeadPlayerIdx(2)
	o.SetTrickNumber(1)
	o.NextTrick()
	assert.Equal(t, domain.NinetyNinePhasePlay, o.GetPhase())
	assert.Equal(t, 2, o.GetCurrentPlayerIdx())
	assert.Equal(t, 2, o.GetTrickNumber())

	o.SetPhase(domain.NinetyNinePhaseBid)
	o.NextTrick() // wrong phase no-op
	assert.Equal(t, domain.NinetyNinePhaseBid, o.GetPhase())
}

// --- Scoring + bonus ---

func setupScore(o *domain.NinetyNine, bids, tricks []int) {
	o.SetPhase(domain.NinetyNinePhaseRoundEnd)
	o.SetDealNumber(1)
	for i := 0; i < 3; i++ {
		p := o.GetPlayer(i)
		p.ResetRound()
		p.SetBid(bids[i])
		for k := 0; k < tricks[i]; k++ {
			p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		}
	}
}

func TestNinetyNine_ScoreRound_AllSucceedBonus(t *testing.T) {
	o := newTestNinetyNine()
	// all three make exactly their bid → each +10 base+bid + bonus 10
	setupScore(o, []int{3, 3, 3}, []int{3, 3, 3})
	o.ScoreRound()
	// 10 + 3 + 10 = 23 each
	for i := 0; i < 3; i++ {
		assert.Equal(t, 23, o.GetPlayer(i).GetCumulativeScore(), "player %d", i)
	}
}

func TestNinetyNine_ScoreRound_SoloSuccessBonus(t *testing.T) {
	o := newTestNinetyNine()
	// only player 0 succeeds → +30 bonus
	setupScore(o, []int{2, 4, 5}, []int{2, 3, 3})
	o.ScoreRound()
	assert.Equal(t, 10+2+30, o.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, o.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 0, o.GetPlayer(2).GetCumulativeScore())
}

func TestNinetyNine_ScoreRound_TwoSucceedBonus(t *testing.T) {
	o := newTestNinetyNine()
	setupScore(o, []int{2, 3, 5}, []int{2, 3, 4})
	o.ScoreRound()
	assert.Equal(t, 10+2+20, o.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 10+3+20, o.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 0, o.GetPlayer(2).GetCumulativeScore())
}

func TestNinetyNine_ScoreRound_GameEndAndWinner(t *testing.T) {
	o := newTestNinetyNine()
	cfg := o.GetConfig()
	cfg.TargetScore = 20
	o.SetConfig(cfg)
	setupScore(o, []int{2, 4, 5}, []int{2, 3, 3}) // p0 solo +42
	o.ScoreRound()
	assert.True(t, o.GetGameEndFlag())
	assert.Equal(t, domain.NinetyNinePhaseGameEnd, o.GetPhase())
	assert.Equal(t, 0, o.GetWinnerIdx())
}

func TestNinetyNine_ScoreRound_WrongPhase(t *testing.T) {
	o := newTestNinetyNine()
	o.SetPhase(domain.NinetyNinePhaseBid)
	o.ScoreRound()
	assert.False(t, o.GetGameEndFlag())
}

func TestNinetyNine_Tiebreak(t *testing.T) {
	o := newTestNinetyNine()
	cfg := o.GetConfig()
	cfg.TargetScore = 10
	o.SetConfig(cfg)
	// p1 and p2 both succeed solo-style at equal cumulative; deterministic
	// Give p1 and p2 equal everything so seat-index breaks the tie.
	o.SetPhase(domain.NinetyNinePhaseRoundEnd)
	for i := 0; i < 3; i++ {
		p := o.GetPlayer(i)
		p.ResetRound()
	}
	// both player1 and player2 reach the same score; player0 fails
	o.GetPlayer(0).SetBid(9)
	o.GetPlayer(1).SetBid(2)
	o.GetPlayer(2).SetBid(2)
	for k := 0; k < 2; k++ {
		o.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		o.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	}
	o.ScoreRound()
	require.True(t, o.GetGameEndFlag())
	// p1 and p2 tie on cumulative & round score → lower seat (1) wins
	assert.Equal(t, 1, o.GetWinnerIdx())
}

// --- NextRound ---

func TestNinetyNine_NextRound(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	o.SetPhase(domain.NinetyNinePhaseRoundEnd)
	o.NextRound()
	assert.Equal(t, 2, o.GetDealNumber())
	assert.Equal(t, domain.NinetyNinePhaseBid, o.GetPhase())
	// deal 2 trump = heart
	assert.Equal(t, domain.CardDesignHeart, o.GetTrumpSuit())

	o.SetPhase(domain.NinetyNinePhasePlay)
	o.NextRound() // wrong phase no-op
	assert.Equal(t, 2, o.GetDealNumber())
}

// --- Hint ---

func TestNinetyNine_GetHint_BidPhase(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	o.SetBidPlayerIdx(0)
	hint := o.GetHint()
	require.NotNil(t, hint)
	assert.Len(t, hint.BuryIndices, 3)
	assert.Equal(t, "strategic_bury", hint.Reason)
}

func TestNinetyNine_GetHint_PlayPhase(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetCurrentPlayerIdx(0)
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	o.GetPlayer(0).SetBid(1)
	hint := o.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestNinetyNine_GetHint_NotHumanTurn(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetCurrentPlayerIdx(1)
	o.SetPhase(domain.NinetyNinePhasePlay)
	assert.Nil(t, o.GetHint())
}

// --- Accessors ---

func TestNinetyNine_Accessors(t *testing.T) {
	o := newTestNinetyNine()
	o.SetPhase(domain.NinetyNinePhasePlay)
	assert.Equal(t, domain.NinetyNinePhasePlay, o.GetPhase())
	o.SetDealNumber(3)
	assert.Equal(t, 3, o.GetDealNumber())
	o.SetTrickNumber(4)
	assert.Equal(t, 4, o.GetTrickNumber())
	o.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, o.GetCurrentPlayerIdx())
	o.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, o.GetLeadPlayerIdx())
	o.SetBidPlayerIdx(2)
	assert.Equal(t, 2, o.GetBidPlayerIdx())
	o.SetDealerIdx(1)
	assert.Equal(t, 1, o.GetDealerIdx())
	o.SetTrumpSuit(domain.CardDesignHeart)
	assert.Equal(t, domain.CardDesignHeart, o.GetTrumpSuit())
	assert.Equal(t, 9, o.GetHandSize())
	assert.Equal(t, 100, o.GetTargetScore())
	assert.Nil(t, o.GetPlayer(99))
	o.Reset()
	o.SetBidPlayerIdx(1)
	o.CpuBid() // generate at least one action-log entry
	assert.NotEmpty(t, o.GetActionLog())
	o.SetCurrentTrick([]*domain.TrickCard{})
	assert.NotNil(t, o.GetCurrentTrick())

	// IsHumanTurn both branches
	o.SetCurrentPlayerIdx(0)
	assert.True(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(1)
	assert.False(t, o.IsHumanTurn())
	o.SetCurrentPlayerIdx(-1)
	assert.False(t, o.IsHumanTurn())

	// IsHumanBidTurn both branches
	o.SetBidPlayerIdx(0)
	assert.True(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(1)
	assert.False(t, o.IsHumanBidTurn())
	o.SetBidPlayerIdx(-1)
	assert.False(t, o.IsHumanBidTurn())
}

func TestNinetyNine_GetValidPlayIndices(t *testing.T) {
	o := newTestNinetyNine()
	setupPlay(o)
	o.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
	})
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	o.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	valid := o.GetValidPlayIndices(0)
	assert.Equal(t, []int{0}, valid) // only the heart
}

// --- JSON round trip + unmarshal hardening ---

func TestNinetyNine_JSON_RoundTrip(t *testing.T) {
	o := newTestNinetyNine()
	o.Reset()
	data, err := json.Marshal(o)
	require.NoError(t, err)

	o2 := &domain.NinetyNine{}
	require.NoError(t, json.Unmarshal(data, o2))
	assert.Equal(t, o.GetDealNumber(), o2.GetDealNumber())
	assert.Equal(t, o.GetTrumpSuit(), o2.GetTrumpSuit())
	assert.Equal(t, o.GetPlayerCnt(), o2.GetPlayerCnt())
}

func validNinetyNineJSON(t *testing.T) map[string]any {
	t.Helper()
	o := newTestNinetyNine()
	o.Reset()
	data, err := json.Marshal(o)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	return m
}

func TestNinetyNine_UnmarshalRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name   string
		tamper func(m map[string]any)
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 99 }},
		{"trumpSuit out of range", func(m map[string]any) { m["ts"] = 99 }},
		{"dealerIdx out of range", func(m map[string]any) { m["di"] = 5 }},
		{"currentPlayerIdx out of range", func(m map[string]any) { m["ci"] = 5 }},
		{"bidPlayerIdx out of range", func(m map[string]any) { m["bi"] = 5 }},
		{"leadPlayerIdx out of range", func(m map[string]any) { m["li"] = 5 }},
		{"trickNumber negative", func(m map[string]any) { m["tn"] = -1 }},
		{"invalid config", func(m map[string]any) { m["cf"] = map[string]any{"cd": 99, "ts": 100} }},
		{"winner invalid when ended", func(m map[string]any) {
			m["ge"] = true
			m["wi"] = 9
		}},
		{"winner out of range not ended", func(m map[string]any) { m["wi"] = 9 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validNinetyNineJSON(t)
			tt.tamper(m)
			data, err := json.Marshal(m)
			require.NoError(t, err)
			o := &domain.NinetyNine{}
			assert.Error(t, json.Unmarshal(data, o))
		})
	}
}

func TestNinetyNine_UnmarshalRejectsBadPlayerCount(t *testing.T) {
	m := validNinetyNineJSON(t)
	m["ps"] = []any{} // 0 players
	data, _ := json.Marshal(m)
	o := &domain.NinetyNine{}
	assert.Error(t, json.Unmarshal(data, o))
}

func TestNinetyNine_UnmarshalRejectsOversized(t *testing.T) {
	big := make([]map[string]any, 1001)
	data, _ := json.Marshal(map[string]any{"al": big, "ps": []any{nil, nil, nil}})
	o := &domain.NinetyNine{}
	assert.Error(t, json.Unmarshal(data, o))
}

func TestNinetyNine_UnmarshalRejectsBadBid(t *testing.T) {
	m := validNinetyNineJSON(t)
	players := m["ps"].([]any)
	p0 := players[0].(map[string]any)
	p0["bd"] = 99
	data, _ := json.Marshal(m)
	o := &domain.NinetyNine{}
	assert.Error(t, json.Unmarshal(data, o))
}
