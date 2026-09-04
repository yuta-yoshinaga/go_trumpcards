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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestCatchTenCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CatchTenCuiPresenter)

	t.Run("play phase", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Catch the Ten")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トランプ:")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("i18n error message", func(t *testing.T) {
		origLang := i18n.Lang()
		t.Cleanup(func() { i18n.SetLang(origLang) })

		m, _ := setupCatchTenWebMockWithPlayers()
		errCode := domain.NewDomainErrorCode(domain.ErrInvalidCard, "catchten.errCardIndexOutOfRange", nil)

		i18n.SetLang("ja")
		jaResult := p.Output(m, errCode)
		assert.Contains(t, jaResult, "カードインデックスが範囲外です")
		assert.NotContains(t, jaResult, "out of range")

		i18n.SetLang("en")
		enResult := p.Output(m, errCode)
		assert.Contains(t, enResult, "out of range")
		assert.NotContains(t, enResult, "カードインデックスが範囲外です")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CatchTenPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end team win", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetWinnerTeam").Return(domain.CatchTenDrawTeam)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestCatchTenCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CatchTenCuiPresenter)

	t.Run("with hint", func(t *testing.T) {
		m, players := setupCatchTenWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		cardIdx := 0
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.CatchTenHint{CardIndex: &cardIdx, Reason: "trump_cut"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupCatchTenWebMockWithPlayers()
		m.On("GetHint").Return((*domain.CatchTenHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestCatchTenCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CatchTenCuiPresenter)
	m := setupCatchTenWebMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "test"},
	})
	assert.NotEmpty(t, p.ActionLogOutput(m))
}

var _ interfaces.CatchTenGame = (*interfaces.MockCatchTenGame)(nil)
