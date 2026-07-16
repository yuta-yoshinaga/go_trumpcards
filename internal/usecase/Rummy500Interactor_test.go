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

func TestNewRummy500Interactor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockRummy500Presenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "Rummy500Interactor: g must not be nil", func() {
			usecase.NewRummy500Interactor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockRummy500Game)
		assert.PanicsWithValue(t, "Rummy500Interactor: gp must not be nil", func() {
			usecase.NewRummy500Interactor(gameMock, nil)
		})
	})
}

func TestRummy500Interactor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockRummy500Presenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockRummy500Game)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewRummy500Interactor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestRummy500Interactor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		cfg := domain.Rummy500Config{CpuDifficulty: domain.Rummy500CpuDifficultyHard, PointLimit: 500}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		gameMock := new(interfaces.MockRummy500Game)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		cfg := domain.Rummy500Config{CpuDifficulty: domain.Rummy500CpuDifficulty(-1), PointLimit: 500}
		assert.Equal(t, "err", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestRummy500Interactor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("error propagated", func(t *testing.T) {
		drawErr := errors.New("draw err")
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})
}

func TestRummy500Interactor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard(0))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard", 2).Return(errors.New("bad"))
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard(2))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard(0))
	})
}

func TestRummy500Interactor_Meld(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Meld([]int{0, 1, 2}))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(errors.New("bad"))
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Meld([]int{0, 1, 2}))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(nil)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Meld([]int{0, 1, 2}))
	})
}

func TestRummy500Interactor_Layoff(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Layoff(0, 0, 0))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerLayoff", 1, 2, 3).Return(errors.New("bad"))
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Layoff(1, 2, 3))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerLayoff", 0, 0, 0).Return(nil)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Layoff(0, 0, 0))
	})
}

func TestRummy500Interactor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(errors.New("bad"))
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})
}

func TestRummy500Interactor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
	})
}

func TestRummy500Interactor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockRummy500Presenter)
	gameMock := new(interfaces.MockRummy500Game)
	cfg := domain.DefaultRummy500Config()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewRummy500Interactor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestRummy500Interactor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockRummy500Presenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log output")
	gameMock := new(interfaces.MockRummy500Game)
	ci := usecase.NewRummy500Interactor(gameMock, pMock)
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestRummy500Interactor_runCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("CPU plays until round end", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("Reset").Return()
		// First iter: not ended, not human → cpuPlay then loop
		gameMock.On("GetGameEndFlag").Return(false).Times(2)
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseDraw).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return().Once()
		// Second iter: now RoundEnd → break
		gameMock.On("GetPhase").Return(domain.Rummy500PhaseRoundEnd).Once()

		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})

	t.Run("stops when game ends mid-loop", func(t *testing.T) {
		pMock := new(presenter.MockRummy500Presenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockRummy500Game)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewRummy500Interactor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestRestoreRummy500Interactor(t *testing.T) {
	pMock := new(presenter.MockRummy500Presenter)
	src := domain.NewDefaultRummy500()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)
	ci, err := usecase.RestoreRummy500Interactor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRummy500Interactor_Hint(t *testing.T) {
	pMock := new(presenter.MockRummy500Presenter)
	pMock.On("HintOutput", mock.Anything).Return("hint output")
	gameMock := new(interfaces.MockRummy500Game)
	ci := usecase.NewRummy500Interactor(gameMock, pMock)
	assert.Equal(t, "hint output", ci.Hint())
}
