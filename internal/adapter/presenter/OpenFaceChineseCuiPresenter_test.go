//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func ofcCardP(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

func setupOpenFaceChineseCuiMock() *interfaces.MockOpenFaceChineseGame {
	m := new(interfaces.MockOpenFaceChineseGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetConfig").Return(domain.DefaultOpenFaceChineseConfig())
	m.On("GetPlayerCnt").Return(2)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	human := domain.NewOpenFaceChinesePlayer(true)
	human.SetPending([]*domain.Card{ofcCardP(domain.CardDesignSpade, 13)})
	human.SetFront([]*domain.Card{ofcCardP(domain.CardDesignHeart, 5)})
	cpu := domain.NewOpenFaceChinesePlayer(false)
	m.On("GetPlayer", 0).Return(human)
	m.On("GetPlayer", 1).Return(cpu)
	m.On("GetCurrentCard").Return(ofcCardP(domain.CardDesignSpade, 13))
	return m
}

func TestOpenFaceChineseCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OpenFaceChineseCuiPresenter)

	t.Run("placing phase shows rows and prompt", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		result := p.Output(m, nil)
		// i18n is loaded (ja) in this test build → assert on rendered Japanese text.
		assert.Contains(t, result, "ラウンド")
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OpenFaceChinesePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner with winner", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end draw", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestOpenFaceChineseCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OpenFaceChineseCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return((*domain.OpenFaceChineseHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint back row", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: domain.OpenFaceChineseRowBack, Reason: "strong_back"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint front row", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: domain.OpenFaceChineseRowFront, Reason: "weak_front"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestOpenFaceChineseCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OpenFaceChineseCuiPresenter)
	m := setupOpenFaceChineseCuiMock()
	result := p.ActionLogOutput(m)
	assert.NotNil(t, result)
}
