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

func TestNewHandAndFootInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockHandAndFootPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "HandAndFootInteractor: g must not be nil", func() {
			usecase.NewHandAndFootInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockHandAndFootGame)
		assert.PanicsWithValue(t, "HandAndFootInteractor: gp must not be nil", func() {
			usecase.NewHandAndFootInteractor(gameMock, nil)
		})
	})
}

func TestHandAndFootInteractor_Hint(t *testing.T) {
	pMock := new(presenter.MockHandAndFootPresenter)
	gameMock := new(interfaces.MockHandAndFootGame)
	pMock.On("HintOutput", gameMock).Return("hint output")

	ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
	assert.Equal(t, "hint output", ci.Hint())
	pMock.AssertCalled(t, "HintOutput", gameMock)
}

func TestHandAndFootInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockHandAndFootPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockHandAndFootGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
	result := ci.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
}

func TestHandAndFootInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		cfg := domain.DefaultHandAndFootConfig()
		cfg.CpuDifficulty = domain.HandAndFootCpuDifficultyHard
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		gameMock := new(interfaces.MockHandAndFootGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		cfg := domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficulty(-1), PointLimit: 100, RedCanastasToGoOut: 1, BlackCanastasToGoOut: 1}
		result := ci.ResetWithConfig(cfg)
		assert.Equal(t, "validation error", result)
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestHandAndFootInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("draw error returned to presenter", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromStock()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestHandAndFootInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error returned", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(drawErr)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(nil)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.DrawFromDiscard([]int{0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard", []int{0, 1})
	})
}

func TestHandAndFootInteractor_Meld(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error returned", func(t *testing.T) {
		meldErr := errors.New("meld error")
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, meldErr).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(meldErr)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.Meld([][]int{{0, 1, 2}})
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid meld runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(nil)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseMeld)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.Meld([][]int{{0, 1, 2}})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerMeld", [][]int{{0, 1, 2}})
	})
}

func TestHandAndFootInteractor_SkipMeld(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockHandAndFootPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockHandAndFootGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once()
	gameMock.On("PlayerSkipMeld").Return(nil)
	gameMock.On("GetPhase").Return(domain.HandAndFootPhaseMeld)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
	result := ci.SkipMeld()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "PlayerSkipMeld")
}

func TestHandAndFootInteractor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error returned", func(t *testing.T) {
		discardErr := errors.New("discard error")
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, discardErr).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", 0).Return(discardErr)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.Discard(0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid discard runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDiscard", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.Discard(3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDiscard", 3)
	})
}

func TestHandAndFootInteractor_GoOut(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error returned", func(t *testing.T) {
		goOutErr := errors.New("go out error")
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, goOutErr).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerGoOut").Return(goOutErr)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.GoOut()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("valid goout runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerGoOut").Return(nil)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.GoOut()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerGoOut")
	})
}

func TestHandAndFootInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestHandAndFootInteractor_GetConfig(t *testing.T) {
	pMock := new(presenter.MockHandAndFootPresenter)
	gameMock := new(interfaces.MockHandAndFootGame)
	expected := domain.DefaultHandAndFootConfig()
	gameMock.On("GetConfig").Return(expected)

	ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
	result := ci.GetConfig()
	assert.Equal(t, expected, result)
}

func TestHandAndFootInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockHandAndFootPresenter)
	gameMock := new(interfaces.MockHandAndFootGame)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	pMock.AssertExpectations(t)
}

func TestHandAndFootInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseRoundEnd)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockHandAndFootPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockHandAndFootGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.HandAndFootPhaseDraw)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewHandAndFootInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}
