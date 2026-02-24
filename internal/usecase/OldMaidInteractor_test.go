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

func newTestOldMaid() *domain.OldMaid {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	return domain.NewOldMaid(domain.NewTrumpCards(1), players)
}

func TestNewOldMaidInteractor_NilGuards(t *testing.T) {
	ompMock := new(presenter.MockOldMaidPresenter)
	t.Run("panics when om is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OldMaidInteractor: om must not be nil", func() {
			usecase.NewOldMaidInteractor(nil, ompMock)
		})
	})
	t.Run("panics when omp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "OldMaidInteractor: omp must not be nil", func() {
			usecase.NewOldMaidInteractor(newTestOldMaid(), nil)
		})
	})
}

func TestOldMaidInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	toi := usecase.NewOldMaidInteractor(newTestOldMaid(), ompMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, toi.Reset())
	})
	t.Run("success Draw", func(t *testing.T) {
		assert.Equal(t, mockOutput, toi.Draw(-1))
	})
}

func TestOldMaidInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOldMaidGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuDraw").Return(nil)
	gameMock.On("PlayerDraw", mock.Anything).Return(nil)

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := oi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Draw calls game.PlayerDraw when human turn", func(t *testing.T) {
		result := oi.Draw(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDraw", 0)
	})
}
