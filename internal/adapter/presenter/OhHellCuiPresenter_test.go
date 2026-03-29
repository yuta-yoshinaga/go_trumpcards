//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupOhHellCuiMock() *interfaces.MockOhHellGame {
	m := new(interfaces.MockOhHellGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTotalRounds").Return(19)
	m.On("GetHandSize").Return(10)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.OhHellTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OhHellPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetRestrictedBid").Return(-1)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultOhHellConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupOhHellCuiMockWithPlayers() (*interfaces.MockOhHellGame, []*domain.OhHellPlayer) {
	m := setupOhHellCuiMock()
	players := makeOhHellPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestOhHellCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.OhHellCuiPresenter)

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Oh Hell")
		assert.Contains(t, result, "ラウンド: 1/19")
		assert.Contains(t, result, "手札枚数: 10")
		assert.Contains(t, result, "切り札:")
		assert.Contains(t, result, "手番:")
	})

	t.Run("bid phase", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ")
	})

	t.Run("bid phase with restriction", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRestrictedBid")
		m.On("GetPhase").Return(domain.OhHellPhaseBid)
		m.On("GetRestrictedBid").Return(3)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッド3は不可")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OhHellPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("no trump", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpCard")
		m.On("GetTrumpCard").Return((*domain.Card)(nil))
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: なし")
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupOhHellCuiMockWithPlayers()
		result := p.Output(m, domain.ErrInvalidPlay)
		assert.Contains(t, result, "invalid play")
	})
}

func TestOhHellCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.OhHellCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		m := setupOhHellCuiMock()
		bid := 3
		m.On("GetHint").Return(&domain.OhHellHint{Bid: &bid, Reason: "strategic_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ビッド 3")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupOhHellCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		cardIdx := 0
		m.On("GetHint").Return(&domain.OhHellHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
	})

	t.Run("nil hint", func(t *testing.T) {
		m := setupOhHellCuiMock()
		m.On("GetHint").Return((*domain.OhHellHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestOhHellCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OhHellCuiPresenter)
	m := setupOhHellCuiMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bids 3"},
	})
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
