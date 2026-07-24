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

func TestNewBurracoInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockBurracoPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BurracoInteractor: g must not be nil", func() {
			usecase.NewBurracoInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBurracoGame)
		assert.PanicsWithValue(t, "BurracoInteractor: gp must not be nil", func() {
			usecase.NewBurracoInteractor(gameMock, nil)
		})
	})
}

func TestBurracoInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("reset calls game Reset and returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
}

func TestBurracoInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		cfg := domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyHard, PointLimit: 5000}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		gameMock := new(interfaces.MockBurracoGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		cfg := domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficulty(-1), PointLimit: 100}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestBurracoInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("draw error is returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestBurracoInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard(nil)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromDiscard")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard(nil)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(drawErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard", []int{0, 1})
	})
}

func TestBurracoInteractor_Meld(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Meld(nil)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerMeld")
	})

	t.Run("error is returned", func(t *testing.T) {
		meldErr := errors.New("meld error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, meldErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(meldErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Meld([][]int{{0, 1, 2}})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid meld runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseMeld)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Meld([][]int{{0, 1, 2}})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerMeld", [][]int{{0, 1, 2}})
	})
}

func TestBurracoInteractor_SkipMeld(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.SkipMeld()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerSkipMeld")
	})

	t.Run("error is returned", func(t *testing.T) {
		skipErr := errors.New("skip error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, skipErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerSkipMeld").Return(skipErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.SkipMeld()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid skip runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerSkipMeld").Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseMeld)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.SkipMeld()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerSkipMeld")
	})
}

func TestBurracoInteractor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDiscard")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("error is returned", func(t *testing.T) {
		discardErr := errors.New("discard error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, discardErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(discardErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid discard runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDiscard", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.Discard(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDiscard", 3)
	})
}

func TestBurracoInteractor_GoOut(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.GoOut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerGoOut")
	})

	t.Run("error is returned", func(t *testing.T) {
		goOutErr := errors.New("go out error")
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, goOutErr).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerGoOut").Return(goOutErr)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.GoOut()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid goout runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerGoOut").Return(nil)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.GoOut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerGoOut")
	})
}

func TestBurracoInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestBurracoInteractor_GetConfig(t *testing.T) {
	t.Run("returns config from game", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		gameMock := new(interfaces.MockBurracoGame)
		expected := domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyHard, PointLimit: 5000}
		gameMock.On("GetConfig").Return(expected)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		result := ci.GetConfig()
		assert.Equal(t, expected, result)
	})
}

func TestBurracoInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockBurracoPresenter)
	gameMock := new(interfaces.MockBurracoGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewBurracoInteractor(gameMock, pMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	pMock.AssertExpectations(t)
}

func TestBurracoInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockBurracoPresenter)
	gameMock := new(interfaces.MockBurracoGame)
	pMock.On("HintOutput", gameMock).Return("recommended: draw from stock")

	ci := usecase.NewBurracoInteractor(gameMock, pMock)
	result := ci.Hint()
	assert.Equal(t, "recommended: draw from stock", result)
	pMock.AssertExpectations(t)
}

func TestBurracoInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when game ended", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseRoundEnd)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is GameEnd", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseGameEnd)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when human turn", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockBurracoPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBurracoGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BurracoPhaseDraw)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBurracoInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}
