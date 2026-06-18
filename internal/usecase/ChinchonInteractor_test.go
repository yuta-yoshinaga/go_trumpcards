//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const chMockOut = `{"phase":0}`

func chNewPresenterMock() *presenter.MockChinchonPresenter {
	p := new(presenter.MockChinchonPresenter)
	p.On("Output", mock.Anything, mock.Anything).Return(chMockOut)
	return p
}

// chGameMockPlayable wires the minimal getters so guardNotPlayable + runCpuTurns
// treat the game as a live human turn.
func chGameMockPlayable() *interfaces.MockChinchonGame {
	m := new(interfaces.MockChinchonGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ChinchonPhaseDraw)
	m.On("IsHumanTurn").Return(true)
	return m
}

func TestNewChinchonInteractor_NilGuards(t *testing.T) {
	pMock := chNewPresenterMock()
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ChinchonInteractor: g must not be nil", func() {
			usecase.NewChinchonInteractor(nil, pMock)
		})
	})
	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockChinchonGame)
		assert.PanicsWithValue(t, "ChinchonInteractor: gp must not be nil", func() {
			usecase.NewChinchonInteractor(gameMock, nil)
		})
	})
}

func TestChinchonInteractor_Reset(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := chGameMockPlayable()
	gameMock.On("Reset").Return()
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, chMockOut, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestChinchonInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid config sets then resets", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("Reset").Return()
		gameMock.On("SetConfig", mock.Anything).Return()
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		out := ci.ResetWithConfig(domain.DefaultChinchonConfig())
		assert.Equal(t, chMockOut, out)
		gameMock.AssertCalled(t, "SetConfig", domain.DefaultChinchonConfig())
	})
	t.Run("invalid config returns error output without reset", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := new(interfaces.MockChinchonGame)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		bad := domain.ChinchonConfig{CpuDifficulty: 99, PlayerCount: 4, KnockThreshold: 5, EliminationLimit: 100}
		out := ci.ResetWithConfig(bad)
		assert.Equal(t, chMockOut, out)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

func TestChinchonInteractor_Actions(t *testing.T) {
	t.Run("DrawFromStock success", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("DrawFromStock error", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.DrawFromStock())
	})
	t.Run("DrawFromDiscard success", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDrawFromDiscard").Return(nil)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.DrawFromDiscard())
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard")
	})
	t.Run("DrawFromDiscard error", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDrawFromDiscard").Return(errors.New("boom"))
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.DrawFromDiscard())
	})
	t.Run("Discard success", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDiscard", 2).Return(nil)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Discard(2))
		gameMock.AssertCalled(t, "PlayerDiscard", 2)
	})
	t.Run("Discard error", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerDiscard", 0).Return(errors.New("boom"))
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Discard(0))
	})
	t.Run("Knock success", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerKnock", 3).Return(nil)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Knock(3))
		gameMock.AssertCalled(t, "PlayerKnock", 3)
	})
	t.Run("Knock error", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerKnock", 0).Return(errors.New("boom"))
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Knock(0))
	})
	t.Run("Layoff success", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerLayoff", mock.Anything).Return(nil)
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Layoff([]int{0, 1}))
		gameMock.AssertCalled(t, "PlayerLayoff", []int{0, 1})
	})
	t.Run("Layoff error", func(t *testing.T) {
		pMock := chNewPresenterMock()
		gameMock := chGameMockPlayable()
		gameMock.On("PlayerLayoff", mock.Anything).Return(errors.New("boom"))
		ci := usecase.NewChinchonInteractor(gameMock, pMock)
		assert.Equal(t, chMockOut, ci.Layoff(nil))
	})
}

func TestChinchonInteractor_NextRound(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := chGameMockPlayable()
	gameMock.On("NextRound").Return()
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, chMockOut, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestChinchonInteractor_NextRound_GameEnded(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := new(interfaces.MockChinchonGame)
	gameMock.On("GetGameEndFlag").Return(true)
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, chMockOut, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestChinchonInteractor_GameEndedGuard(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := new(interfaces.MockChinchonGame)
	gameMock.On("GetGameEndFlag").Return(true)
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, chMockOut, ci.DrawFromStock())
	gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
}

func TestChinchonInteractor_GetConfig(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := new(interfaces.MockChinchonGame)
	gameMock.On("GetConfig").Return(domain.DefaultChinchonConfig())
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, domain.DefaultChinchonConfig(), ci.GetConfig())
}

func TestChinchonInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockChinchonPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log-out")
	gameMock := new(interfaces.MockChinchonGame)
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	assert.Equal(t, "log-out", ci.ActionLog())
}

func TestChinchonInteractor_RunCpuTurns(t *testing.T) {
	pMock := chNewPresenterMock()
	gameMock := new(interfaces.MockChinchonGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ChinchonPhaseDiscard)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	ci := usecase.NewChinchonInteractor(gameMock, pMock)
	ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreChinchonInteractor(t *testing.T) {
	g := domain.NewDefaultChinchon()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)
	pMock := chNewPresenterMock()
	ci, err := usecase.RestoreChinchonInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}
