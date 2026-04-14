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

func newTestFiftyOneGame() *domain.FiftyOne {
	players := []*domain.FiftyOnePlayer{
		domain.NewFiftyOnePlayer(true),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
	}
	return domain.NewFiftyOne(domain.NewTrumpCards(0), players)
}

func TestNewFiftyOneInteractor_NilGuards(t *testing.T) {
	fopMock := new(presenter.MockFiftyOnePresenter)
	t.Run("panics when fo is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "FiftyOneInteractor: fo must not be nil", func() {
			usecase.NewFiftyOneInteractor(nil, fopMock)
		})
	})
	t.Run("panics when fop is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "FiftyOneInteractor: fop must not be nil", func() {
			usecase.NewFiftyOneInteractor(newTestFiftyOneGame(), nil)
		})
	})
}

func TestFiftyOneInteractor_Reset(t *testing.T) {
	mockOutput := `{"players":[]}`
	fopMock := new(presenter.MockFiftyOnePresenter)
	fopMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	fi := usecase.NewFiftyOneInteractor(newTestFiftyOneGame(), fopMock)

	t.Run("success with default config", func(t *testing.T) {
		result := fi.Reset(domain.DefaultFiftyOneConfig())
		assert.Equal(t, mockOutput, result)
	})
	t.Run("invalid config returns error output", func(t *testing.T) {
		cfg := domain.FiftyOneConfig{CpuDifficulty: domain.FiftyOneCpuDifficulty(-1)}
		result := fi.Reset(cfg)
		assert.Equal(t, mockOutput, result)
	})
}

func TestFiftyOneInteractor_MockGame(t *testing.T) {
	mockOutput := `{"ok":true}`
	fopMock := new(presenter.MockFiftyOnePresenter)
	fopMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	fopMock.On("ActionLogOutput", mock.Anything).Return(`{"log":[]}`)

	gameMock := new(interfaces.MockFiftyOneGame)
	gameMock.On("Reset").Return()
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return(nil)
	gameMock.On("ExchangeOne", mock.Anything, mock.Anything).Return(nil)
	gameMock.On("ExchangeAll").Return(nil)
	gameMock.On("Stop").Return(nil)
	gameMock.On("GetConfig").Return(domain.DefaultFiftyOneConfig())

	fi := usecase.NewFiftyOneInteractor(gameMock, fopMock)

	t.Run("Reset calls SetConfig and game.Reset", func(t *testing.T) {
		result := fi.Reset(domain.DefaultFiftyOneConfig())
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("ExchangeOne calls game.ExchangeOne", func(t *testing.T) {
		result := fi.ExchangeOne(0, 1)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ExchangeOne", 0, 1)
	})

	t.Run("ExchangeAll calls game.ExchangeAll", func(t *testing.T) {
		result := fi.ExchangeAll()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ExchangeAll")
	})

	t.Run("Stop calls game.Stop", func(t *testing.T) {
		result := fi.Stop()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Stop")
	})

	t.Run("GetConfig returns game config", func(t *testing.T) {
		cfg := fi.GetConfig()
		assert.Equal(t, domain.DefaultFiftyOneConfig(), cfg)
	})

	t.Run("ActionLog returns action log output", func(t *testing.T) {
		result := fi.ActionLog()
		assert.Equal(t, `{"log":[]}`, result)
	})
}

func TestFiftyOneInteractor_GameEnded(t *testing.T) {
	mockOutput := `{"ended":true}`
	fopMock := new(presenter.MockFiftyOnePresenter)
	fopMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockFiftyOneGame)
	gameMock.On("GetGameEndFlag").Return(true)

	fi := usecase.NewFiftyOneInteractor(gameMock, fopMock)

	t.Run("ExchangeOne blocked", func(t *testing.T) {
		result := fi.ExchangeOne(0, 0)
		assert.Equal(t, mockOutput, result)
	})
	t.Run("ExchangeAll blocked", func(t *testing.T) {
		result := fi.ExchangeAll()
		assert.Equal(t, mockOutput, result)
	})
	t.Run("Stop blocked", func(t *testing.T) {
		result := fi.Stop()
		assert.Equal(t, mockOutput, result)
	})
}

func TestFiftyOneInteractor_SnapshotRestore(t *testing.T) {
	fo := newTestFiftyOneGame()
	fo.Reset()

	fopMock := new(presenter.MockFiftyOnePresenter)
	fopMock.On("Output", mock.Anything, mock.Anything).Return(`{}`)
	fi := usecase.NewFiftyOneInteractor(fo, fopMock)

	data, err := fi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreFiftyOneInteractor(data, fopMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
