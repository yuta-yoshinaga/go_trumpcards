//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestCallBreak() *domain.CallBreak {
	players := []*domain.CallBreakPlayer{
		domain.NewCallBreakPlayer(true),
		domain.NewCallBreakPlayer(false),
		domain.NewCallBreakPlayer(false),
		domain.NewCallBreakPlayer(false),
	}
	return domain.NewCallBreak(domain.NewTrumpCards(0), players, domain.DefaultCallBreakConfig())
}

func setupCallBreakBid(cb *domain.CallBreak, idx int) {
	cb.SetPhase(domain.CallBreakPhaseBid)
	cb.SetBidPlayerIdx(idx)
}

func setupCallBreakPlay(cb *domain.CallBreak, current, lead, trickNum int) {
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(current)
	cb.SetLeadPlayerIdx(lead)
	cb.SetTrickNumber(trickNum)
}

func TestNewCallBreak(t *testing.T) {
	cb := newTestCallBreak()
	assert.Equal(t, -1, cb.GetWinnerIdx())
	assert.Equal(t, 0, cb.GetRoundNumber())
}

func TestNewDefaultCallBreak(t *testing.T) {
	cb := domain.NewDefaultCallBreak()
	require.NotNil(t, cb)
	assert.Equal(t, domain.CallBreakPlayerCnt, cb.GetPlayerCnt())
	assert.True(t, cb.GetPlayer(0).GetIsHuman())
	for i := 1; i < cb.GetPlayerCnt(); i++ {
		assert.False(t, cb.GetPlayer(i).GetIsHuman())
	}
	assert.False(t, cb.GetGameEndFlag())
	assert.Equal(t, domain.DefaultCallBreakConfig(), cb.GetConfig())
}

func TestCallBreak_Reset(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()

	assert.Equal(t, domain.CallBreakPhaseBid, cb.GetPhase())
	assert.Equal(t, 1, cb.GetRoundNumber())
	assert.Equal(t, 0, cb.GetTrickNumber())
	assert.False(t, cb.GetSpadesBroken())
	assert.False(t, cb.GetGameEndFlag())
	assert.Equal(t, -1, cb.GetWinnerIdx())
	assert.Equal(t, 0, cb.GetBidPlayerIdx())
	assert.Equal(t, 0, cb.GetLeadPlayerIdx())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, cb.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, cb.GetPlayer(i).GetBid())
		assert.Equal(t, 0, cb.GetPlayer(i).GetCumulativeScore())
	}
}

func TestCallBreak_Reset_ClearsAccumulated(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()

	cb.GetPlayer(0).SetCumulativeScore(123)
	cb.SetPhase(domain.CallBreakPhaseGameEnd)

	cb.Reset()
	assert.Equal(t, 0, cb.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, domain.CallBreakPhaseBid, cb.GetPhase())
}

func TestCallBreak_PlayerBid_Valid(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakBid(cb, 0)
	err := cb.PlayerBid(3)
	require.NoError(t, err)
	assert.Equal(t, 3, cb.GetPlayer(0).GetBid())
}

func TestCallBreak_PlayerBid_MaxBid(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakBid(cb, 0)
	err := cb.PlayerBid(domain.CallBreakHandSize)
	require.NoError(t, err)
	assert.Equal(t, domain.CallBreakHandSize, cb.GetPlayer(0).GetBid())
}

func TestCallBreak_PlayerBid_BelowMin_Errors(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakBid(cb, 0)
	err := cb.PlayerBid(0) // Nil ビッドは不可
	assert.Error(t, err)
}

func TestCallBreak_PlayerBid_AboveMax_Errors(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakBid(cb, 0)
	err := cb.PlayerBid(14)
	assert.Error(t, err)
}

func TestCallBreak_PlayerBid_WrongPhase(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhasePlay)
	err := cb.PlayerBid(3)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestCallBreak_PlayerBid_NotHumanTurn(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetBidPlayerIdx(1) // CPU
	err := cb.PlayerBid(3)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestCallBreak_CpuBid_AdvancesIndex(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetBidPlayerIdx(1)
	cb.CpuBid()
	assert.GreaterOrEqual(t, cb.GetPlayer(1).GetBid(), domain.CallBreakMinBid)
	assert.Equal(t, 2, cb.GetBidPlayerIdx())
}

func TestCallBreak_CpuBid_SkipsHuman(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetBidPlayerIdx(0) // Human
	cb.CpuBid()           // 何もしない
	assert.Equal(t, -1, cb.GetPlayer(0).GetBid())
	assert.Equal(t, 0, cb.GetBidPlayerIdx())
}

func TestCallBreak_AllBidsTransitionToPlay(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()

	require.NoError(t, cb.PlayerBid(3))
	for i := 1; i < domain.CallBreakPlayerCnt; i++ {
		cb.CpuBid()
	}
	assert.Equal(t, domain.CallBreakPhasePlay, cb.GetPhase())
	assert.Equal(t, 1, cb.GetTrickNumber())
}

func TestCallBreak_ScoreRound_BidExact(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhaseRoundEnd)

	for i := 0; i < domain.CallBreakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(3)
	}
	// プレイヤー 0 はぴったり 3 取った: 30
	for k := 0; k < 3; k++ {
		cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	// プレイヤー 1 は 5 取った: bid 3 → 32 (overtricks=2)
	for k := 0; k < 5; k++ {
		cb.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	// プレイヤー 2 は 1 しか取れず: -30
	cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	// プレイヤー 3 は 4 取った: 31
	for k := 0; k < 4; k++ {
		cb.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}

	cb.ScoreRound()

	assert.Equal(t, 30, cb.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 32, cb.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, -30, cb.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 31, cb.GetPlayer(3).GetCumulativeScore())
}

func TestCallBreak_ScoreRound_WrongPhase_NoOp(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.GetPlayer(0).SetBid(3)
	cb.ScoreRound()
	assert.Equal(t, 0, cb.GetPlayer(0).GetCumulativeScore())
}

func TestCallBreak_ScoreRound_TriggersGameEndAtMaxRounds(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cfg := cb.GetConfig()
	cfg.MaxRounds = 2
	cb.SetConfig(cfg)

	// 1st round
	cb.SetPhase(domain.CallBreakPhaseRoundEnd)
	for i := 0; i < domain.CallBreakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(2)
		for k := 0; k < 2; k++ {
			cb.GetPlayer(i).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
	}
	cb.ScoreRound()
	assert.False(t, cb.GetGameEndFlag(), "should not end after round 1 of 2")

	cb.NextRound()
	require.Equal(t, 2, cb.GetRoundNumber())

	// 2nd / final round
	cb.SetPhase(domain.CallBreakPhaseRoundEnd)
	for i := 0; i < domain.CallBreakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(2)
	}
	// プレイヤー 2 が一番たくさん取って勝者
	for k := 0; k < 2; k++ {
		cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	for k := 0; k < 2; k++ {
		cb.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	for k := 0; k < 5; k++ {
		cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	for k := 0; k < 2; k++ {
		cb.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}

	cb.ScoreRound()

	assert.True(t, cb.GetGameEndFlag())
	assert.Equal(t, domain.CallBreakPhaseGameEnd, cb.GetPhase())
	assert.Equal(t, 2, cb.GetWinnerIdx(), "player 2 should win with highest cumulative score")
}

func TestCallBreak_NextRound_OnlyFromRoundEnd(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhasePlay)
	before := cb.GetRoundNumber()
	cb.NextRound()
	assert.Equal(t, before, cb.GetRoundNumber())
}

func TestCallBreak_NextRound_AdvancesAndDealsAgain(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhaseRoundEnd)
	cb.NextRound()
	assert.Equal(t, 2, cb.GetRoundNumber())
	assert.Equal(t, domain.CallBreakPhaseBid, cb.GetPhase())
	for i := 0; i < domain.CallBreakPlayerCnt; i++ {
		assert.Equal(t, 13, cb.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, cb.GetPlayer(i).GetBid())
	}
}

func TestCallBreak_NextTrick_OnlyFromTrickEnd(t *testing.T) {
	cb := newTestCallBreak()
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.SetTrickNumber(3)
	cb.NextTrick()
	assert.Equal(t, 3, cb.GetTrickNumber())
}

func TestCallBreak_NextTrick_Advances(t *testing.T) {
	cb := newTestCallBreak()
	cb.SetPhase(domain.CallBreakPhaseTrickEnd)
	cb.SetLeadPlayerIdx(2)
	cb.SetTrickNumber(3)
	cb.NextTrick()
	assert.Equal(t, 4, cb.GetTrickNumber())
	assert.Equal(t, 2, cb.GetCurrentPlayerIdx())
	assert.Equal(t, domain.CallBreakPhasePlay, cb.GetPhase())
}

func TestCallBreak_PlayerPlay_WrongPhase(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhaseBid)
	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestCallBreak_PlayerPlay_NotHumanTurn(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakPlay(cb, 1, 1, 1)
	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestCallBreak_PlayerPlay_InvalidIndex(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	setupCallBreakPlay(cb, 0, 0, 1)
	err := cb.PlayerPlay(99)
	assert.Error(t, err)
}

func TestCallBreak_PlayerPlay_GameEnded(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()

	cfg := cb.GetConfig()
	cfg.MaxRounds = 1
	cb.SetConfig(cfg)

	cb.SetPhase(domain.CallBreakPhaseRoundEnd)
	for i := 0; i < domain.CallBreakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(1)
		cb.GetPlayer(i).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	cb.ScoreRound()
	require.True(t, cb.GetGameEndFlag())

	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestCallBreak_ResolveTrick(t *testing.T) {
	cb := newTestCallBreak()
	cb.SetPhase(domain.CallBreakPhaseTrickEnd)
	cb.SetTrickNumber(1)
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	}
	cb.SetCurrentTrick(trick)
	cb.ResolveTrick()
	assert.Equal(t, 1, cb.GetPlayer(1).GetTrickCount())
	assert.Equal(t, 1, cb.GetLeadPlayerIdx())
}

func TestCallBreak_ResolveTrick_TrumpWins(t *testing.T) {
	cb := newTestCallBreak()
	cb.SetPhase(domain.CallBreakPhaseTrickEnd)
	cb.SetTrickNumber(1)
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	}
	cb.SetCurrentTrick(trick)
	cb.ResolveTrick()
	assert.Equal(t, 1, cb.GetPlayer(1).GetTrickCount())
}

func TestCallBreak_ResolveTrick_AdvancesToRoundEnd(t *testing.T) {
	cb := newTestCallBreak()
	cb.SetPhase(domain.CallBreakPhaseTrickEnd)
	cb.SetTrickNumber(domain.CallBreakHandSize)
	cb.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})
	cb.ResolveTrick()
	assert.Equal(t, domain.CallBreakPhaseRoundEnd, cb.GetPhase())
}

func TestCallBreak_IsHumanTurn(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()

	cb.SetCurrentPlayerIdx(-1)
	assert.False(t, cb.IsHumanTurn())

	cb.SetCurrentPlayerIdx(0)
	assert.True(t, cb.IsHumanTurn())

	cb.SetCurrentPlayerIdx(1)
	assert.False(t, cb.IsHumanTurn())
}

func TestCallBreak_IsHumanBidTurn(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetBidPlayerIdx(0)
	assert.True(t, cb.IsHumanBidTurn())
	cb.SetBidPlayerIdx(2)
	assert.False(t, cb.IsHumanBidTurn())
	cb.SetBidPlayerIdx(99)
	assert.False(t, cb.IsHumanBidTurn())
}

func TestCallBreak_GetPlayer_OutOfRange(t *testing.T) {
	cb := newTestCallBreak()
	assert.Nil(t, cb.GetPlayer(-1))
	assert.Nil(t, cb.GetPlayer(99))
}

func TestCallBreak_GetHint_BidPhase(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	hint := cb.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
	assert.Nil(t, hint.CardIndex)
	assert.NotEmpty(t, hint.Reason)
}

func TestCallBreak_GetHint_PlayPhase(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.SetCurrentTrick(nil)
	cb.SetSpadesBroken(true)
	hint := cb.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestCallBreak_GetHint_NotPlayerTurn(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(1)
	assert.Nil(t, cb.GetHint())
}

func TestCallBreak_GetValidPlayIndices_NotEmpty(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	indices := cb.GetValidPlayIndices(0)
	assert.NotEmpty(t, indices)
}

func TestCallBreak_GetActionLog(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	require.NoError(t, cb.PlayerBid(3))
	log := cb.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestFormatCallBreakScore(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{41, "4.1"},
		{30, "3.0"},
		{-40, "-4.0"},
		{0, "0.0"},
		{132, "13.2"},
		{-132, "-13.2"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.FormatCallBreakScore(c.in))
	}
}

func TestCallBreak_JSONRoundTrip(t *testing.T) {
	cb := newTestCallBreak()
	cb.Reset()
	require.NoError(t, cb.PlayerBid(4))

	data, err := json.Marshal(cb)
	require.NoError(t, err)

	cb2 := newTestCallBreak()
	require.NoError(t, json.Unmarshal(data, cb2))
	assert.Equal(t, cb.GetPhase(), cb2.GetPhase())
	assert.Equal(t, cb.GetRoundNumber(), cb2.GetRoundNumber())
	assert.Equal(t, 4, cb2.GetPlayer(0).GetBid())
}

func TestCallBreak_UnmarshalJSON_Invalid(t *testing.T) {
	cb := newTestCallBreak()
	err := json.Unmarshal([]byte("not json"), cb)
	assert.Error(t, err)
}

func TestCallBreak_UnmarshalJSON_Empty(t *testing.T) {
	cb := newTestCallBreak()
	require.NoError(t, json.Unmarshal([]byte("{}"), cb))
	assert.Empty(t, cb.GetCurrentTrick())
	assert.Empty(t, cb.GetActionLog())
}
