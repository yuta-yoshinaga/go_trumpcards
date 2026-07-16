package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupGongZhuCuiMock() *interfaces.MockGongZhuGame {
	m := new(interfaces.MockGongZhuGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetExposure").Return(domain.GongZhuExposure{})
	m.On("GetCurrentTrick").Return([]*domain.GongZhuTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GongZhuPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupGongZhuCuiMockWithPlayers() (*interfaces.MockGongZhuGame, []*domain.GongZhuPlayer) {
	m := setupGongZhuCuiMock()
	players := makeGongZhuPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGongZhuCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GongZhuCuiPresenter)

	t.Run("play phase shows player info and prompt", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Gong Zhu")
		assert.NotEmpty(t, result)
	})

	t.Run("captured point cards are listed per player", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		// CPU 1 took a trick containing the pig (♠Q), a heart, and a plain card.
		players[1].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 12, false),  // pig -> shown
			domain.NewCard(domain.CardDesignHeart, 5, false),   // heart -> shown
			domain.NewCard(domain.CardDesignClover, 3, false),  // plain -> not shown
			domain.NewCard(domain.CardDesignDiamond, 8, false), // plain -> not shown
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "獲得: SPADE 12 HEART 5")
		// Only the one capturing player gets a line; the others took nothing.
		assert.Equal(t, 1, strings.Count(result, "獲得:"))
	})

	t.Run("expose phase prompt", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposure")
		m.On("GetPhase").Return(domain.GongZhuPhaseExpose)
		m.On("GetExposure").Return(domain.GongZhuExposure{Pig: true, Sheep: true, Ace: true, Doubler: true})
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// Exposure summary uses the localized point-card symbol keys for every exposed card.
		assert.Contains(t, result, i18n.T("gongzhu.card.spadeQueen"))
		assert.Contains(t, result, i18n.T("gongzhu.card.diamondJack"))
		assert.Contains(t, result, i18n.T("gongzhu.card.heartAce"))
		assert.Contains(t, result, i18n.T("gongzhu.card.clubTen"))
	})

	t.Run("trick end and round end prompts", func(t *testing.T) {
		for _, phase := range []domain.GongZhuPhase{domain.GongZhuPhaseTrickEnd, domain.GongZhuPhaseRoundEnd} {
			m, _ := setupGongZhuCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			assert.NotEmpty(t, p.Output(m, nil))
		}
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestGongZhuCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GongZhuCuiPresenter)

	t.Run("expose hint", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{0}, Reason: "expose_sheep"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("expose none hint (empty indices)", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{}, Reason: "expose_none"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.GongZhuHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestGongZhuCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GongZhuCuiPresenter)
	m := new(interfaces.MockGongZhuGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
