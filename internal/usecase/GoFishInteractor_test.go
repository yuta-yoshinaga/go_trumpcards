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

func newTestGoFishGame() *domain.GoFish {
	players := []*domain.GoFishPlayer{
		domain.NewGoFishPlayer(true),
		domain.NewGoFishPlayer(false),
		domain.NewGoFishPlayer(false),
		domain.NewGoFishPlayer(false),
	}
	return domain.NewGoFish(domain.NewTrumpCards(0), players)
}

func TestNewGoFishInteractor_NilGuards(t *testing.T) {
	gfpMock := new(presenter.MockGoFishPresenter)
	t.Run("panics when gf is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GoFishInteractor: gf must not be nil", func() {
			usecase.NewGoFishInteractor(nil, gfpMock)
		})
	})
	t.Run("panics when gfp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GoFishInteractor: gfp must not be nil", func() {
			usecase.NewGoFishInteractor(newTestGoFishGame(), nil)
		})
	})
}

func TestGoFishInteractor_Reset(t *testing.T) {
	mockOutput := `{"players":[]}`
	gfpMock := new(presenter.MockGoFishPresenter)
	gfpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gi := usecase.NewGoFishInteractor(newTestGoFishGame(), gfpMock)

	t.Run("success with default config", func(t *testing.T) {
		result := gi.Reset(domain.DefaultGoFishConfig())
		assert.Equal(t, mockOutput, result)
	})
	t.Run("invalid config returns error output", func(t *testing.T) {
		cfg := domain.GoFishConfig{CpuDifficulty: domain.GoFishCpuDifficulty(-1)}
		result := gi.Reset(cfg)
		assert.Equal(t, mockOutput, result)
	})
}

func TestGoFishInteractor_MockGame(t *testing.T) {
	mockOutput := `{"ok":true}`
	gfpMock := new(presenter.MockGoFishPresenter)
	gfpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gfpMock.On("ActionLogOutput", mock.Anything).Return(`{"log":[]}`)

	gameMock := new(interfaces.MockGoFishGame)
	gameMock.On("Reset").Return()
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuAsk").Return(nil)
	gameMock.On("PlayerAsk", mock.Anything, mock.Anything).Return(nil)
	gameMock.On("GetConfig").Return(domain.DefaultGoFishConfig())

	gi := usecase.NewGoFishInteractor(gameMock, gfpMock)

	t.Run("Reset calls SetConfig and game.Reset", func(t *testing.T) {
		result := gi.Reset(domain.DefaultGoFishConfig())
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Ask calls game.PlayerAsk", func(t *testing.T) {
		result := gi.Ask(1, 3)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerAsk", 1, 3)
	})

	t.Run("GetConfig returns game config", func(t *testing.T) {
		cfg := gi.GetConfig()
		assert.Equal(t, domain.DefaultGoFishConfig(), cfg)
	})

	t.Run("ActionLog returns action log output", func(t *testing.T) {
		result := gi.ActionLog()
		assert.Equal(t, `{"log":[]}`, result)
	})
}

func TestGoFishInteractor_Ask_GameEnded(t *testing.T) {
	mockOutput := `{"ended":true}`
	gfpMock := new(presenter.MockGoFishPresenter)
	gfpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockGoFishGame)
	gameMock.On("GetGameEndFlag").Return(true)

	gi := usecase.NewGoFishInteractor(gameMock, gfpMock)
	result := gi.Ask(1, 3)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerAsk")
}

func TestGoFishInteractor_Ask_NotHumanTurn(t *testing.T) {
	mockOutput := `{"cpu_turn":true}`
	gfpMock := new(presenter.MockGoFishPresenter)
	gfpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockGoFishGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	gi := usecase.NewGoFishInteractor(gameMock, gfpMock)
	result := gi.Ask(1, 3)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerAsk")
}

func TestGoFishInteractor_SnapshotRestore(t *testing.T) {
	gf := newTestGoFishGame()
	gf.Reset()

	gfpMock := new(presenter.MockGoFishPresenter)
	gfpMock.On("Output", mock.Anything, mock.Anything).Return(`{}`)
	gi := usecase.NewGoFishInteractor(gf, gfpMock)

	data, err := gi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreGoFishInteractor(data, gfpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
