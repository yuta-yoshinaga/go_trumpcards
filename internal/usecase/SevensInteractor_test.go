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

func newTestSevens() *domain.Sevens {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	return domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
}

func TestNewSevensInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSevensPresenter)
	t.Run("panics when s is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SevensInteractor: s must not be nil", func() {
			usecase.NewSevensInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SevensInteractor: sp must not be nil", func() {
			usecase.NewSevensInteractor(newTestSevens(), nil)
		})
	})
}

func TestSevensInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tsi := usecase.NewSevensInteractor(newTestSevens(), spMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Reset())
	})

	t.Run("success ResetWithConfig all enabled", func(t *testing.T) {
		result := tsi.ResetWithConfig(true, 2, true, domain.SevensMaxPasses)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig default values", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 0, false, domain.SevensMaxPasses)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig jokerCount clamped to 0", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, -5, false, domain.SevensMaxPasses)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig jokerCount clamped to 2", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 10, false, domain.SevensMaxPasses)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig maxPasses 0 (unlimited)", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 0, false, 0)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig maxPasses 3", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 0, false, 3)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success ResetWithConfig maxPasses negative clamped to 0", func(t *testing.T) {
		result := tsi.ResetWithConfig(false, 0, false, -1)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("success Play with pass (idx -1)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Play(-1))
	})

	t.Run("success Play with index", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.Play(0))
	})

	t.Run("success PlayJoker", func(t *testing.T) {
		assert.Equal(t, mockOutput, tsi.PlayJoker(0, 1, 6))
	})
}

func TestSevensInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	spMock := new(presenter.MockSevensPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSevensGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("HasAnyOption", mock.Anything).Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("PlayerPlayJoker", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	gameMock.On("GetCurrentTurn").Return(0)

	si := usecase.NewSevensInteractor(gameMock, spMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := si.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Play calls game.PlayerPlay when human turn", func(t *testing.T) {
		result := si.Play(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", 0)
	})
	t.Run("PlayJoker calls game.PlayerPlayJoker when human turn", func(t *testing.T) {
		result := si.PlayJoker(0, 1, 6)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlayJoker", 0, 1, 6)
	})
}
