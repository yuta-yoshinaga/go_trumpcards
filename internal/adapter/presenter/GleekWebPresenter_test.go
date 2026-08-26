//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGleekWebMock() *interfaces.MockGleekGame {
	m := new(interfaces.MockGleekGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GleekPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetCurrentBidderIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(2)
	m.On("GetElderIdx").Return(0)
	m.On("GetBuyerIdx").Return(0)
	m.On("GetWinningBid").Return(14)
	m.On("HighestBid").Return(14)
	m.On("NextBidAmount").Return(16)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTurnUp").Return(domain.NewCard(domain.CardDesignHeart, 4, false))
	m.On("GetRuffWinnerIdx").Return(0)
	m.On("GetRuffs").Return([]*domain.GleekRuff{
		{PlayerIdx: 0, Suit: domain.CardDesignHeart, Total: 31},
		{PlayerIdx: 1, Suit: domain.CardDesignSpade, Total: 24},
		{PlayerIdx: 2, Suit: domain.CardDesignClover, Total: 20},
	})
	m.On("GetMelds").Return([]*domain.GleekMeld{{PlayerIdx: 0, Rank: 13, Count: 3, Value: 3}})
	m.On("GetTrickPoints").Return([domain.GleekPlayerCnt]int{9, 3, 0})
	m.On("GetLastTrickWinner").Return(1)
	m.On("DealPoints").Return(78)
	m.On("Par").Return(26)
	m.On("GetBids").Return([domain.GleekPlayerCnt]int{14, 12, 0})
	m.On("GetPassed").Return([domain.GleekPlayerCnt]bool{false, true, true})
	m.On("GetResult").Return(domain.GleekResultNone)
	m.On("GetPlayerScores").Return([domain.GleekPlayerCnt]int{})
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayableIndices", 0).Return([]int{0})
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("IsHumanDiscardTurn").Return(false)
	m.On("GetConfig").Return(domain.DefaultGleekConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetHint").Return((*domain.GleekHint)(nil)).Maybe()
	return m
}

func setupGleekWebMockWithPlayers() (*interfaces.MockGleekGame, []*domain.GleekPlayer) {
	m := setupGleekWebMock()
	players := makeGleekPlayers()
	m.On("GetPlayerCnt").Return(domain.GleekPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func decodeGleek(t *testing.T, raw string) *controller.GleekWebOutput {
	t.Helper()
	out := new(controller.GleekWebOutput)
	require.NoError(t, json.Unmarshal([]byte(raw), out))
	return out
}

func TestGleekWebPresenter_Output(t *testing.T) {
	p := new(presenter.GleekWebPresenter)

	t.Run("initial state carries the stage results", func(t *testing.T) {
		m, players := setupGleekWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := decodeGleek(t, p.Output(m, nil))

		assert.Len(t, out.Players, domain.GleekPlayerCnt)
		assert.Equal(t, int(domain.GleekPhasePlay), out.Phase)
		assert.Equal(t, 0, out.BuyerIdx)
		assert.Equal(t, 14, out.WinningBid)
		assert.Equal(t, 16, out.NextBidAmount)
		assert.Equal(t, domain.GleekSwapSize, out.DiscardCount)
		assert.Equal(t, 0, out.RuffWinnerIdx)
		assert.Equal(t, 78, out.DealPoints)
		assert.Equal(t, 26, out.Par)
		// **直前のトリックを取った席を運ぶ。** -1 を固定で返していると、画面は
		// 誰が取ったかを出せないまま次へ進むボタンだけを見せる。
		assert.Equal(t, 1, out.LastTrickWinner)
		require.Len(t, out.Melds, 1)
		assert.Equal(t, 13, out.Melds[0].Rank)
		assert.Equal(t, 3, out.Melds[0].Value)
		require.NotNil(t, out.TurnUp)
		assert.Equal(t, "gleek.playPhase.lead", out.MessageCode)
	})

	// **段階の読みは席ごとに載る。** ラフとスートを落とすと、画面は誰が
	// 何で取ったのか出せない。
	t.Run("each seat carries its bid, ruff and trick points", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		out := decodeGleek(t, p.Output(m, nil))
		assert.True(t, out.Players[0].IsBuyer)
		assert.False(t, out.Players[1].IsBuyer)
		assert.Equal(t, 14, out.Players[0].Bid)
		assert.True(t, out.Players[1].Passed)
		assert.Equal(t, 31, out.Players[0].Ruff)
		assert.Equal(t, domain.CardDesignHeart, out.Players[0].RuffSuit)
		assert.Equal(t, 9, out.Players[0].TrickPoints)
	})

	t.Run("phase message codes", func(t *testing.T) {
		for _, tc := range []struct {
			phase domain.GleekPhase
			code  string
		}{
			{domain.GleekPhaseBid, "gleek.bidPhase"},
			{domain.GleekPhaseDiscard, "gleek.discardPhase"},
			{domain.GleekPhaseTrickEnd, "gleek.trickEnd"},
			{domain.GleekPhaseRoundEnd, "gleek.roundEnd"},
		} {
			m, _ := setupGleekWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(tc.phase)
			assert.Equal(t, tc.code, decodeGleek(t, p.Output(m, nil)).MessageCode)
		}
	})

	t.Run("following a lead uses the follow message", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
		})
		assert.Equal(t, "gleek.playPhase.follow", decodeGleek(t, p.Output(m, nil)).MessageCode)
	})

	t.Run("errors ride on the message, not a code", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		out := decodeGleek(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})

	t.Run("game end names the winner side", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		assert.Equal(t, "gleek.result.humanWin", decodeGleek(t, p.Output(m, nil)).MessageCode)

		m2, _ := setupGleekWebMockWithPlayers()
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetGameEndFlag")
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetWinnerPlayer")
		m2.On("GetGameEndFlag").Return(true)
		m2.On("GetWinnerPlayer").Return(1)
		out := decodeGleek(t, p.Output(m2, nil))
		assert.Equal(t, "gleek.result.cpuWin", out.MessageCode)
		assert.Equal(t, "1", out.MessageParams["player"])
	})

	// **押していない人にヒントを渡さない側は Output、押した側は HintOutput。**
	t.Run("hint output flags that the hint was requested", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.GleekHint{CardIndices: []int{1}, Reason: "lead_high"})
		out := decodeGleek(t, p.HintOutput(m))
		assert.Equal(t, "gleek.hintRequested", out.MessageCode)
		require.NotNil(t, out.Hint)
		assert.Equal(t, []int{1}, out.Hint.CardIndices)

		m2, _ := setupGleekWebMockWithPlayers()
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetHint")
		m2.On("GetHint").Return((*domain.GleekHint)(nil))
		assert.Equal(t, "gleek.noHint", decodeGleek(t, p.HintOutput(m2)).MessageCode)
	})

	t.Run("playable indices are empty outside the human play turn", func(t *testing.T) {
		m, _ := setupGleekWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(false)
		assert.Empty(t, decodeGleek(t, p.Output(m, nil)).PlayableIndices)
	})
}
