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

func TestNewCatchTenInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockCatchTenPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CatchTenInteractor: g must not be nil", func() {
			usecase.NewCatchTenInteractor(nil, cpMock)
		})
	})

	t.Run("panics when cp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCatchTenGame)
		assert.PanicsWithValue(t, "CatchTenInteractor: cp must not be nil", func() {
			usecase.NewCatchTenInteractor(gameMock, nil)
		})
	})
}

func TestCatchTenInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	cpMock := new(presenter.MockCatchTenPresenter)
	cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCatchTenGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CatchTenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestCatchTenInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	cpMock := new(presenter.MockCatchTenPresenter)
	cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCatchTenGame)
	cfg := domain.CatchTenConfig{CpuDifficulty: domain.CatchTenCpuDifficultyHard, PointLimit: 50}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CatchTenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestCatchTenInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	cpMock := new(presenter.MockCatchTenPresenter)
	gameMock := new(interfaces.MockCatchTenGame)
	cpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	invalidCfg := domain.CatchTenConfig{CpuDifficulty: 99, PointLimit: 41}
	assert.Equal(t, "validation error", ci.ResetWithConfig(invalidCfg))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestCatchTenInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid play", func(t *testing.T) {
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.CatchTenPhasePlay)

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, mockOutput, ci.Play(2))
		gameMock.AssertCalled(t, "PlayerPlay", 2)
	})

	t.Run("play error", func(t *testing.T) {
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return("error output")
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 0).Return(errors.New("invalid"))

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, "error output", ci.Play(0))
	})

	t.Run("game ended guard", func(t *testing.T) {
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return("ended")
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, "ended", ci.Play(0))
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("human completes trick resolves it", func(t *testing.T) {
		// When the human plays the last card of a trick, the interactor
		// resolves it (a CPU-completed trick is resolved in runCpuTurns).
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.CatchTenPhaseTrickEnd)
		gameMock.On("ResolveTrick").Return()

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, mockOutput, ci.Play(1))
		gameMock.AssertCalled(t, "ResolveTrick")
	})
}

func TestCatchTenInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":0}`
	cpMock := new(presenter.MockCatchTenPresenter)
	cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCatchTenGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CatchTenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, mockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestCatchTenInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("scores round then starts next", func(t *testing.T) {
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.CatchTenPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "ScoreRound")
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended after scoring", func(t *testing.T) {
		cpMock := new(presenter.MockCatchTenPresenter)
		cpMock.On("Output", mock.Anything, mock.Anything).Return("game ended")
		gameMock := new(interfaces.MockCatchTenGame)
		gameMock.On("ScoreRound").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
		assert.Equal(t, "game ended", ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestCatchTenInteractor_GetConfig(t *testing.T) {
	cfg := domain.DefaultCatchTenConfig()
	gameMock := new(interfaces.MockCatchTenGame)
	gameMock.On("GetConfig").Return(cfg)
	cpMock := new(presenter.MockCatchTenPresenter)

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestCatchTenInteractor_Hint(t *testing.T) {
	cpMock := new(presenter.MockCatchTenPresenter)
	gameMock := new(interfaces.MockCatchTenGame)
	cpMock.On("HintOutput", gameMock).Return("hint output")

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, "hint output", ci.Hint())
}

func TestCatchTenInteractor_ActionLog(t *testing.T) {
	cpMock := new(presenter.MockCatchTenPresenter)
	gameMock := new(interfaces.MockCatchTenGame)
	cpMock.On("ActionLogOutput", gameMock).Return("log output")

	ci := usecase.NewCatchTenInteractor(gameMock, cpMock)
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestRestoreCatchTenInteractor(t *testing.T) {
	g := domain.NewDefaultCatchTen()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	cpMock := new(presenter.MockCatchTenPresenter)
	ci, err := usecase.RestoreCatchTenInteractor(data, cpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}
