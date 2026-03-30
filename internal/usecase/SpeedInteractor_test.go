//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestSpeed() *domain.Speed {
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	return domain.NewSpeed(domain.NewTrumpCards(0), players, domain.DefaultSpeedConfig())
}

func TestNewSpeedInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSpeedPresenter)

	t.Run("panics when s is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SpeedInteractor: s must not be nil", func() {
			usecase.NewSpeedInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SpeedInteractor: sp must not be nil", func() {
			usecase.NewSpeedInteractor(newTestSpeed(), nil)
		})
	})
}

func TestSpeedInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSpeedPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSpeedGame)
	gameMock.On("Reset").Return()

	si := usecase.NewSpeedInteractor(gameMock, spMock)
	result := si.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
}

func TestSpeedInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets", func(t *testing.T) {
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpeedGame)
		cfg := domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("invalid config returns error output", func(t *testing.T) {
		errOutput := `{"error":"invalid"}`
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOutput)
		gameMock := new(interfaces.MockSpeedGame)
		cfg := domain.SpeedConfig{CpuDifficulty: -1}

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.ResetWithConfig(cfg)
		assert.Equal(t, errOutput, result)
	})
}

func TestSpeedInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid play then CPU responds", func(t *testing.T) {
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPlay", 0, 1).Return(nil)
		gameMock.On("CpuPlay").Return([]*domain.SpeedCpuAction{})
		gameMock.On("UpdatePhase").Return()

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Play(0, 1)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 0, 1)
		gameMock.AssertCalled(t, "CpuPlay")
		gameMock.AssertCalled(t, "UpdatePhase")
	})

	t.Run("invalid play returns error", func(t *testing.T) {
		errOutput := `{"error":"invalid"}`
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOutput)
		spMock.On("Output", mock.Anything, nil).Return(mockOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerPlay", 0, 0).Return(domain.ErrInvalidPlay)

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Play(0, 0)
		assert.Equal(t, errOutput, result)
	})

	t.Run("game ended blocks play", func(t *testing.T) {
		endOutput := `{"gameEnd":true}`
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Play(0, 0)
		assert.Equal(t, endOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("play causes game end skips CPU", func(t *testing.T) {
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(false).Once()
		gameMock.On("PlayerPlay", 0, 0).Return(nil)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("UpdatePhase").Return()

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Play(0, 0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestSpeedInteractor_Flip(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("successful flip then CPU plays", func(t *testing.T) {
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Flip").Return(nil)
		gameMock.On("CpuPlay").Return([]*domain.SpeedCpuAction{})
		gameMock.On("UpdatePhase").Return()

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Flip()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Flip")
		gameMock.AssertCalled(t, "CpuPlay")
	})

	t.Run("flip error", func(t *testing.T) {
		errOutput := `{"error":"wrong phase"}`
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.MatchedBy(func(err error) bool { return err != nil })).Return(errOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("Flip").Return(domain.ErrWrongPhase)

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Flip()
		assert.Equal(t, errOutput, result)
	})

	t.Run("game ended blocks flip", func(t *testing.T) {
		endOutput := `{"gameEnd":true}`
		spMock := new(presenter.MockSpeedPresenter)
		spMock.On("Output", mock.Anything, mock.Anything).Return(endOutput)
		gameMock := new(interfaces.MockSpeedGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewSpeedInteractor(gameMock, spMock)
		result := si.Flip()
		assert.Equal(t, endOutput, result)
		gameMock.AssertNotCalled(t, "Flip")
	})
}

func TestSpeedInteractor_Hint(t *testing.T) {
	mockOutput := `{"hint":true}`
	spMock := new(presenter.MockSpeedPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSpeedGame)

	si := usecase.NewSpeedInteractor(gameMock, spMock)
	result := si.Hint()
	assert.Equal(t, mockOutput, result)
}

func TestSpeedInteractor_GetConfig(t *testing.T) {
	cfg := domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard}
	gameMock := new(interfaces.MockSpeedGame)
	gameMock.On("GetConfig").Return(cfg)
	spMock := new(presenter.MockSpeedPresenter)

	si := usecase.NewSpeedInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
}

func TestSpeedInteractor_ActionLog(t *testing.T) {
	logOutput := `{"log":[]}`
	spMock := new(presenter.MockSpeedPresenter)
	spMock.On("ActionLogOutput", mock.Anything).Return(logOutput)
	gameMock := new(interfaces.MockSpeedGame)

	si := usecase.NewSpeedInteractor(gameMock, spMock)
	result := si.ActionLog()
	assert.Equal(t, logOutput, result)
}

func TestSpeedInteractor_Snapshot(t *testing.T) {
	game := newTestSpeed()
	game.Reset()
	spMock := new(presenter.MockSpeedPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	si := usecase.NewSpeedInteractor(game, spMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreSpeedInteractor(t *testing.T) {
	game := newTestSpeed()
	game.Reset()
	spMock := new(presenter.MockSpeedPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return("{}")

	si := usecase.NewSpeedInteractor(game, spMock)
	data, _ := si.Snapshot()

	restored, err := usecase.RestoreSpeedInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreSpeedInteractor_InvalidJSON(t *testing.T) {
	spMock := new(presenter.MockSpeedPresenter)
	_, err := usecase.RestoreSpeedInteractor([]byte("invalid"), spMock)
	assert.Error(t, err)
}
