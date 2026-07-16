//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupWizardCuiMock() *interfaces.MockWizardGame {
	m := new(interfaces.MockWizardGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTotalRounds").Return(15)
	m.On("GetHandSize").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.WizardTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WizardPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetRestrictedBid").Return(-1)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultWizardConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupWizardCuiMockWithPlayers() (*interfaces.MockWizardGame, []*domain.WizardPlayer) {
	m := setupWizardCuiMock()
	players := makeWizardPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWizardCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.WizardCuiPresenter)

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Wizard")
		assert.Contains(t, result, "ラウンド: 1/15")
		assert.Contains(t, result, "手札枚数: 1")
		assert.Contains(t, result, "切り札:")
		assert.Contains(t, result, "手番:")
		// Empty trick → no established lead suit yet.
		assert.Contains(t, result, i18n.T("wizard.leadNone"))
	})

	t.Run("play phase names the lead suit", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		// A Jester (skipped) then a heart establishes hearts as the lead suit.
		m.On("GetCurrentTrick").Return([]*domain.WizardTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.WizardDesignJester, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "HEART")
		assert.NotContains(t, result, i18n.T("wizard.leadNone"))
	})

	t.Run("renders wizard and jester cards in hand", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.WizardDesignWizard, 1, false))
		players[0].AddCard(domain.NewCard(domain.WizardDesignJester, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Wizard")
		assert.Contains(t, result, "Jester")
	})

	t.Run("bid phase", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ")
	})

	t.Run("bid phase with restriction", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRestrictedBid")
		m.On("GetPhase").Return(domain.WizardPhaseBid)
		m.On("GetRestrictedBid").Return(3)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッド3は不可")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("no trump", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpCard")
		m.On("GetTrumpCard").Return((*domain.Card)(nil))
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: なし")
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		result := p.Output(m, domain.ErrInvalidPlay)
		assert.Contains(t, result, "invalid play")
	})
}

func TestWizardCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.WizardCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		m := setupWizardCuiMock()
		bid := 3
		m.On("GetHint").Return(&domain.WizardHint{Bid: &bid, Reason: "strategic_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ビッド 3")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		cardIdx := 0
		m.On("GetHint").Return(&domain.WizardHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
	})

	t.Run("nil hint", func(t *testing.T) {
		m := setupWizardCuiMock()
		m.On("GetHint").Return((*domain.WizardHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint with nil bid and nil cardIndex", func(t *testing.T) {
		m := setupWizardCuiMock()
		m.On("GetHint").Return(&domain.WizardHint{Reason: "unknown"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("card hint with no human player", func(t *testing.T) {
		m := setupWizardCuiMock()
		cpuPlayers := []*domain.WizardPlayer{
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
		}
		m.On("GetPlayerCnt").Return(4)
		m.On("GetPlayer", 0).Return(cpuPlayers[0])
		m.On("GetPlayer", 1).Return(cpuPlayers[1])
		m.On("GetPlayer", 2).Return(cpuPlayers[2])
		m.On("GetPlayer", 3).Return(cpuPlayers[3])
		cardIdx := 0
		m.On("GetHint").Return(&domain.WizardHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestWizardCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WizardCuiPresenter)
	m := setupWizardCuiMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bids 3"},
	})
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
