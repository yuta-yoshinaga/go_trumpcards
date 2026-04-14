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

func TestFiftyOneWebPresenter_Output_Initial(t *testing.T) {
	p := new(presenter.FiftyOneWebPresenter)
	m := setupFiftyOneMock()

	result := p.Output(m, nil)
	var out controller.FiftyOneWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.Len(t, out.Players, 4)
	assert.Equal(t, 0, out.CurrentTurn)
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Len(t, out.TableCards, 5)
	assert.Equal(t, -1, out.StopCallerIdx)
}

func TestFiftyOneWebPresenter_Output_GameEnd_HumanWin(t *testing.T) {
	p := new(presenter.FiftyOneWebPresenter)
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.FiftyOnePhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetTurnNumber").Return(10)
	m.On("GetStopCallerIdx").Return(0)
	m.On("GetLastAction").Return("stop")
	m.On("GetLastHandIdx").Return(-1)
	m.On("GetLastTableIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m.On("GetTableCards").Return([]*domain.Card{})

	humanPlayer := domain.NewFiftyOnePlayer(true)
	humanPlayer.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	m.On("GetPlayer", 0).Return(humanPlayer)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewFiftyOnePlayer(false))
	}

	result := p.Output(m, nil)
	var out controller.FiftyOneWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.True(t, out.GameEndFlag)
	assert.Equal(t, "fiftyone.result.humanWin", out.MessageCode)
}

func TestFiftyOneWebPresenter_Output_GameEnd_CpuWin(t *testing.T) {
	p := new(presenter.FiftyOneWebPresenter)
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.FiftyOnePhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(2)
	m.On("GetTurnNumber").Return(10)
	m.On("GetStopCallerIdx").Return(2)
	m.On("GetLastAction").Return("stop")
	m.On("GetLastHandIdx").Return(-1)
	m.On("GetLastTableIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	m.On("GetTableCards").Return([]*domain.Card{})

	for i := range 4 {
		m.On("GetPlayer", i).Return(domain.NewFiftyOnePlayer(i == 0))
	}

	result := p.Output(m, nil)
	var out controller.FiftyOneWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.True(t, out.GameEndFlag)
	assert.Equal(t, "fiftyone.result.cpuWin", out.MessageCode)
	assert.Equal(t, "2", out.MessageParams["cpuId"])
}

func TestFiftyOneWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.FiftyOneWebPresenter)
	m := setupFiftyOneMock()

	result := p.Output(m, errors.New("bad input"))
	var out controller.FiftyOneWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.Equal(t, "error", out.MessageCode)
	assert.Contains(t, out.Message, "bad input")
}

func TestFiftyOneWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.FiftyOneWebPresenter)
	m := new(interfaces.MockFiftyOneGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
