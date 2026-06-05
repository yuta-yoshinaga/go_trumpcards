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

func TestNewYanivInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockYanivPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "YanivInteractor: g must not be nil", func() {
			usecase.NewYanivInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockYanivGame)
		assert.PanicsWithValue(t, "YanivInteractor: gp must not be nil", func() {
			usecase.NewYanivInteractor(gameMock, nil)
		})
	})
}

func TestYanivInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockYanivPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockYanivGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.YanivPhaseDiscard)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewYanivInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestYanivInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		cfg := domain.YanivConfig{CpuDifficulty: domain.YanivCpuDifficultyHard, ScoreLimit: 100}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.YanivPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		gameMock := new(interfaces.MockYanivGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		cfg := domain.YanivConfig{CpuDifficulty: domain.YanivCpuDifficulty(-1), ScoreLimit: 200}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestYanivInteractor_Discard(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDiscard", []int{0, 1}).Return(nil)
		gameMock.On("GetPhase").Return(domain.YanivPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard([]int{0, 1}))
		gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 1})
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDiscard", []int{0}).Return(err)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard([]int{0}))
	})

	t.Run("not human turn", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard([]int{0}))
		gameMock.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	})
}

func TestYanivInteractor_DeclareYaniv(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareYaniv").Return(nil)
		gameMock.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DeclareYaniv())
		gameMock.AssertCalled(t, "PlayerDeclareYaniv")
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDeclareYaniv").Return(err)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DeclareYaniv())
	})
}

func TestYanivInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.YanivPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(err)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
}

func TestYanivInteractor_DrawFromPickup(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDrawFromPickup", 1).Return(nil)
		gameMock.On("GetPhase").Return(domain.YanivPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromPickup(1))
		gameMock.AssertCalled(t, "PlayerDrawFromPickup", 1)
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("e")
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, err).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromPickup", 0).Return(err)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromPickup(0))
	})
}

func TestYanivInteractor_NextRound(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.YanivPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("game ended", func(t *testing.T) {
		pMock := new(presenter.MockYanivPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockYanivGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewYanivInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

func TestYanivInteractor_GetConfigAndLog(t *testing.T) {
	pMock := new(presenter.MockYanivPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockYanivGame)
	cfg := domain.YanivConfig{CpuDifficulty: domain.YanivCpuDifficultyHard, ScoreLimit: 150}
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewYanivInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestYanivInteractor_RunCpuTurns_ExitsOnRoundEnd(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockYanivPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockYanivGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.YanivPhaseRoundEnd)

	ci := usecase.NewYanivInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestRestoreYanivInteractor(t *testing.T) {
	pMock := new(presenter.MockYanivPresenter)
	g := domain.NewDefaultYaniv()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	ci, err := usecase.RestoreYanivInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreYanivInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockYanivPresenter)
	_, err := usecase.RestoreYanivInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}
