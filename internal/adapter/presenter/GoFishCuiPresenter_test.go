//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestGoFishCuiPresenter_Output_Initial(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Go Fish")
	assert.Contains(t, result, "山札: 32枚")
	assert.Contains(t, result, "のターン")
}

func TestGoFishCuiPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.GoFishPhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetTurnNumber").Return(20)
	m.On("GetDeckRemaining").Return(0)
	m.On("GetConfig").Return(domain.DefaultGoFishConfig())
	m.On("GetLastAskPlayerIdx").Return(-1)
	m.On("GetCpuActions").Return(([]*domain.GoFishCpuAction)(nil))
	m.On("GetHumanAction").Return((*domain.GoFishCpuAction)(nil))

	humanPlayer := domain.NewGoFishPlayer(true)
	m.On("GetPlayer", 0).Return(humanPlayer)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewGoFishPlayer(false))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "あなたの勝ち")
}

func TestGoFishCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()

	result := p.Output(m, domain.ErrGoFishInvalidRank)
	assert.Contains(t, result, "Error:")
	assert.Contains(t, result, "you do not hold that rank")
}

func TestGoFishCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
