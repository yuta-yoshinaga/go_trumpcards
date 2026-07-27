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

func setupGaigelCuiMock() *interfaces.MockGaigelGame {
	m := new(interfaces.MockGaigelGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GaigelPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetStockRemaining").Return(27)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeGaigelPlayers() []*domain.GaigelPlayer {
	return []*domain.GaigelPlayer{
		domain.NewGaigelPlayer(true, 0),
		domain.NewGaigelPlayer(false, 1),
		domain.NewGaigelPlayer(false, 0),
		domain.NewGaigelPlayer(false, 1),
	}
}

func setupGaigelCuiMockWithPlayers() (*interfaces.MockGaigelGame, []*domain.GaigelPlayer) {
	m := setupGaigelCuiMock()
	players := makeGaigelPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGaigelCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.GaigelCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupGaigelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "ガイゲル")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "切り札: SPADE")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("trump undecided", func(t *testing.T) {
		m, _ := setupGaigelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未確定")
	})

	t.Run("marriage available prompt lists candidate cards", func(t *testing.T) {
		m, players := setupGaigelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageIndices")
		m.On("GetMarriageIndices", 0).Return([]int{0, 1})

		result := p.Output(m, nil)
		assert.Contains(t, result, "マリアージュ")
		// The human's K/Q candidate indices are enumerated.
		assert.Contains(t, result, strings.Split(i18n.T("gaigel.promptMarriageCards"), "{{")[0])
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "[1]")
	})

	t.Run("cpu turn does not leak marriage cards", func(t *testing.T) {
		m, players := setupGaigelCuiMockWithPlayers()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageIndices")
		m.On("GetCurrentPlayerIdx").Return(1) // a CPU
		m.On("GetMarriageIndices", 1).Return([]int{0, 1})

		result := p.Output(m, nil)
		// Generic hint may show, but the CPU's specific candidate cards must not.
		assert.NotContains(t, result, strings.Split(i18n.T("gaigel.promptMarriageCards"), "{{")[0])
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupGaigelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GaigelPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック完了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupGaigelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GaigelPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド完了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupGaigelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestGaigelCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.GaigelCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupGaigelCuiMock()
		m.On("GetHint").Return((*domain.GaigelHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupGaigelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.GaigelHint{CardIndex: &idx, Reason: "follow_cut"})
		assert.Contains(t, p.HintOutput(m), "ヒント")
	})

	t.Run("marriage hint", func(t *testing.T) {
		m, players := setupGaigelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		idx := 0
		m.On("GetHint").Return(&domain.GaigelHint{CardIndex: &idx, Reason: "marriage", IsMarriage: true})
		assert.Contains(t, p.HintOutput(m), "マリアージュ")
	})
}

func TestGaigelCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupGaigelCuiMock()
	p := new(presenter.GaigelCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(m))
}
