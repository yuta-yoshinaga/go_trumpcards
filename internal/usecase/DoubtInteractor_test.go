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

func newTestDoubt() *domain.Doubt {
	players := []*domain.DoubtPlayer{
		domain.NewDoubtPlayer(true),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
		domain.NewDoubtPlayer(false),
	}
	return domain.NewDoubt(domain.NewTrumpCards(0), players)
}

func TestNewDoubtInteractor_NilGuards(t *testing.T) {
	dpMock := new(presenter.MockDoubtPresenter)

	t.Run("panics when d is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DoubtInteractor: d must not be nil", func() {
			usecase.NewDoubtInteractor(nil, dpMock)
		})
	})

	t.Run("panics when dp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DoubtInteractor: dp must not be nil", func() {
			usecase.NewDoubtInteractor(newTestDoubt(), nil)
		})
	})
}

func TestDoubtInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	dpMock := new(presenter.MockDoubtPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockDoubtGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.DoubtPhasePlay)

	di := usecase.NewDoubtInteractor(gameMock, dpMock)
	result := di.Reset()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "Reset")
}

func TestDoubtInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("sets config then resets game", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		cfg := domain.DoubtConfig{DoubtWindowSec: 3, CpuMemoryLevel: domain.DoubtMemoryLevelHard}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", cfg)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("default config works", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		cfg := domain.DefaultDoubtConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
	})

	t.Run("with real game - config persists", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := newTestDoubt()
		di := usecase.NewDoubtInteractor(game, dpMock)
		cfg := domain.DoubtConfig{DoubtWindowSec: 5, CpuMemoryLevel: domain.DoubtMemoryLevelEasy}
		result := di.ResetWithConfig(cfg)
		assert.Equal(t, mockOutput, result)
		assert.Equal(t, 5, game.GetConfig().DoubtWindowSec)
		assert.Equal(t, domain.DoubtMemoryLevelEasy, game.GetConfig().CpuMemoryLevel)
	})
}

func TestDoubtInteractor_Play(t *testing.T) {
	mockOutput := `{"phase":1}`

	t.Run("game ended returns output without playing", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetGameEndFlag").Return(true)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.Play([]int{0}, 5)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("not human turn returns output without playing", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.Play([]int{0}, 5)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "PlayerPlay")
	})

	t.Run("valid play calls PlayerPlay", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", []int{0, 1}, 5).Return(nil)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.Play([]int{0, 1}, 5)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", []int{0, 1}, 5)
	})

	t.Run("play error is passed to presenter", func(t *testing.T) {
		playErr := errors.New("test error")
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, playErr).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerPlay", mock.Anything, mock.Anything).Return(playErr)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.Play([]int{0}, 5)
		assert.Equal(t, mockOutput, result)
	})
}

func TestDoubtInteractor_ResolveDoubt(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("calls ResolveDoubt and runs CPU turns until human turn", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("ResolveDoubt", []int{1}).Return()
		// runCpuTurns: not ended, not doubt phase, not human turn → CpuPlay
		// then: not ended, not doubt phase, human turn → break
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.DoubtPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.ResolveDoubt([]int{1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ResolveDoubt", []int{1})
		gameMock.AssertCalled(t, "CpuPlay")
	})

	t.Run("stops when game ended", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("ResolveDoubt", mock.Anything).Return()
		gameMock.On("GetGameEndFlag").Return(true)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.ResolveDoubt([]int{1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("stops when phase is Doubt (CPU played)", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("ResolveDoubt", mock.Anything).Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.DoubtPhaseDoubt)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.ResolveDoubt([]int{1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestDoubtInteractor_SkipDoubt(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("calls SkipDoubt and runs CPU turns", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("SkipDoubt").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.DoubtPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.SkipDoubt()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SkipDoubt")
	})
}

func TestDoubtInteractor_GetCpuDoubters(t *testing.T) {
	t.Run("returns cpu doubters from domain game", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetCpuDoubters").Return([]int{1, 2})

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.GetCpuDoubters()
		assert.Equal(t, []int{1, 2}, result)
	})

	t.Run("returns nil when no cpu doubters", func(t *testing.T) {
		dpMock := new(presenter.MockDoubtPresenter)
		gameMock := new(interfaces.MockDoubtGame)
		gameMock.On("GetCpuDoubters").Return([]int(nil))

		di := usecase.NewDoubtInteractor(gameMock, dpMock)
		result := di.GetCpuDoubters()
		assert.Nil(t, result)
	})
}

func TestDoubtInteractor_WithRealGame(t *testing.T) {
	t.Run("Reset initializes game and returns output", func(t *testing.T) {
		mockOutput := `{"phase":0}`
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		di := usecase.NewDoubtInteractor(newTestDoubt(), dpMock)
		result := di.Reset()
		assert.Equal(t, mockOutput, result)
	})

	t.Run("Play with real game advances state", func(t *testing.T) {
		mockOutput := `{"phase":1}`
		dpMock := new(presenter.MockDoubtPresenter)
		dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		game := newTestDoubt()
		game.Reset()
		// Give player[0] 2 cards
		game.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		game.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		di := usecase.NewDoubtInteractor(game, dpMock)
		// Play with index 0, claimed value 1
		result := di.Play([]int{0}, 1)
		assert.Equal(t, mockOutput, result)
	})
}
