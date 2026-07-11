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

func TestNewZhengInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockZhengPresenter)

	t.Run("panics when zg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ZhengInteractor: zg must not be nil", func() {
			usecase.NewZhengInteractor(nil, pMock)
		})
	})

	t.Run("panics when zp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockZhengGame)
		assert.PanicsWithValue(t, "ZhengInteractor: zp must not be nil", func() {
			usecase.NewZhengInteractor(gameMock, nil)
		})
	})
}

func TestZhengInteractor_Reset(t *testing.T) {
	mockOutput := `{"players":[]}`
	pMock := new(presenter.MockZhengPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockZhengGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)

	zi := usecase.NewZhengInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, zi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestZhengInteractor_Reset_RunsCpuTurns(t *testing.T) {
	mockOutput := `{}`
	pMock := new(presenter.MockZhengPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockZhengGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// CPU turn first, then the human's turn ends the loop.
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	zi := usecase.NewZhengInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, zi.Reset())
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestZhengInteractor_Play(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid play runs cpu turns", func(t *testing.T) {
		pMock := new(presenter.MockZhengPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockZhengGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", []int{0}).Return(nil)
		gameMock.On("HasPendingAction").Return(false)

		zi := usecase.NewZhengInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, zi.Play([]int{0}))
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})

	t.Run("play error surfaces", func(t *testing.T) {
		playErr := errors.New("invalid play")
		pMock := new(presenter.MockZhengPresenter)
		pMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockZhengGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", []int{1}).Return(playErr)

		zi := usecase.NewZhengInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, zi.Play([]int{1}))
	})

	t.Run("game ended blocks play", func(t *testing.T) {
		pMock := new(presenter.MockZhengPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockZhengGame)
		gameMock.On("GetGameEndFlag").Return(true)

		zi := usecase.NewZhengInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, zi.Play([]int{0}))
		gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})
}

func TestZhengInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{}`

	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockZhengPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockZhengGame)
		cfg := domain.ZhengConfig{CpuDifficulty: domain.ZhengDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)

		zi := usecase.NewZhengInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, zi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockZhengPresenter)
		gameMock := new(interfaces.MockZhengGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		zi := usecase.NewZhengInteractor(gameMock, pMock)
		cfg := domain.ZhengConfig{CpuDifficulty: domain.ZhengCpuDifficulty(-1)}
		assert.Equal(t, "validation error", zi.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestZhengInteractor_GetConfigAndLog(t *testing.T) {
	pMock := new(presenter.MockZhengPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockZhengGame)
	cfg := domain.ZhengConfig{CpuDifficulty: domain.ZhengDifficultyHard}
	gameMock.On("GetConfig").Return(cfg)

	zi := usecase.NewZhengInteractor(gameMock, pMock)
	assert.Equal(t, cfg, zi.GetConfig())
	assert.Equal(t, "log", zi.ActionLog())
}

func TestRestoreZhengInteractor(t *testing.T) {
	pMock := new(presenter.MockZhengPresenter)
	g := domain.NewDefaultZheng()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	zi, err := usecase.RestoreZhengInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, zi)
}

func TestRestoreZhengInteractor_InvalidJSON(t *testing.T) {
	pMock := new(presenter.MockZhengPresenter)
	_, err := usecase.RestoreZhengInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}
