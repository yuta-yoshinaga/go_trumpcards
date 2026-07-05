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

const panMockOutput = `{"phase":0}`

func TestNewPanInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockPanPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PanInteractor: g must not be nil", func() {
			usecase.NewPanInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockPanGame)
		assert.PanicsWithValue(t, "PanInteractor: gp must not be nil", func() {
			usecase.NewPanInteractor(gameMock, nil)
		})
	})
}

func TestPanInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockPanPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
	gameMock := new(interfaces.MockPanGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.PanPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewPanInteractor(gameMock, pMock)
	assert.Equal(t, panMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestPanInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		cfg := domain.DefaultPanConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.PanPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		gameMock := new(interfaces.MockPanGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
		ci := usecase.NewPanInteractor(gameMock, pMock)
		cfg := domain.DefaultPanConfig()
		cfg.PlayerCount = 1
		assert.Equal(t, "err", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestPanInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("error propagated", func(t *testing.T) {
		drawErr := errors.New("draw err")
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromStock())
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.PanPhasePlay)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromStock())
	})
}

func TestPanInteractor_DrawFromDiscard(t *testing.T) {
	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromDiscard())
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard").Return(errors.New("bad"))
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromDiscard())
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard").Return(nil)
		gameMock.On("GetPhase").Return(domain.PanPhasePlay)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.DrawFromDiscard())
	})
}

func TestPanInteractor_Meld(t *testing.T) {
	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Meld([]int{0, 1, 2}))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(errors.New("bad"))
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Meld([]int{0, 1, 2}))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", []int{0, 1, 2}).Return(nil)
		gameMock.On("GetPhase").Return(domain.PanPhasePlay)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Meld([]int{0, 1, 2}))
	})
}

func TestPanInteractor_Layoff(t *testing.T) {
	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Layoff(0, 0, 0))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerLayoff", 1, 2, 3).Return(errors.New("bad"))
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Layoff(1, 2, 3))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerLayoff", 0, 0, 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.PanPhasePlay)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Layoff(0, 0, 0))
	})
}

func TestPanInteractor_Discard(t *testing.T) {
	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Discard(0))
	})

	t.Run("error propagated", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(errors.New("bad"))
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Discard(0))
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(nil)
		gameMock.On("GetPhase").Return(domain.PanPhaseDraw)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.Discard(0))
	})
}

func TestPanInteractor_NextRound(t *testing.T) {
	t.Run("blocked when game ended", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("success", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.PanPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		assert.Equal(t, panMockOutput, ci.NextRound())
	})
}

func TestPanInteractor_GetConfigAndActionLog(t *testing.T) {
	pMock := new(presenter.MockPanPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log output")
	gameMock := new(interfaces.MockPanGame)
	cfg := domain.DefaultPanConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewPanInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestPanInteractor_runCpuTurns(t *testing.T) {
	t.Run("CPU plays until round end", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false).Times(2)
		gameMock.On("GetPhase").Return(domain.PanPhaseDraw).Once()
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return().Once()
		gameMock.On("GetPhase").Return(domain.PanPhaseRoundEnd).Once()

		ci := usecase.NewPanInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})

	t.Run("stops when game ends mid-loop", func(t *testing.T) {
		pMock := new(presenter.MockPanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(panMockOutput)
		gameMock := new(interfaces.MockPanGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewPanInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestRestorePanInteractor(t *testing.T) {
	pMock := new(presenter.MockPanPresenter)
	src := domain.NewDefaultPan()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)
	ci, err := usecase.RestorePanInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}
