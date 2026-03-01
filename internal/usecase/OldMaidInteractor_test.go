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

	t.Run("success Reset with DefaultOldMaidConfig", func(t *testing.T) {
		assert.Equal(t, mockOutput, toi.Reset(domain.DefaultOldMaidConfig()))
	})
	t.Run("success Reset with JijiNuki config", func(t *testing.T) {
		cfg := domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki, CpuPlacementStrategy: false}
		assert.Equal(t, mockOutput, toi.Reset(cfg))
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
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("ArrangeTargetForHumanDraw").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuDraw").Return(nil)
	gameMock.On("PlayerDraw", mock.Anything).Return(nil)

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)

	t.Run("Reset calls SetConfig and game.Reset and ArrangeTargetForHumanDraw", func(t *testing.T) {
		result := oi.Reset(domain.DefaultOldMaidConfig())
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertCalled(t, "Reset")
		gameMock.AssertCalled(t, "ArrangeTargetForHumanDraw")
	})
	t.Run("Draw calls game.PlayerDraw and ArrangeTargetForHumanDraw when human turn", func(t *testing.T) {
		result := oi.Draw(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerDraw", 0)
		gameMock.AssertCalled(t, "ArrangeTargetForHumanDraw")
	})
}

func TestOldMaidInteractor_Draw_GameEnded(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOldMaidGame)
	gameMock.On("GetGameEndFlag").Return(true)

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerDraw", mock.Anything)
}

func TestOldMaidInteractor_Draw_NotHumanTurn(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOldMaidGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerDraw", mock.Anything)
}

func TestOldMaidInteractor_Draw_PlayerDrawError(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOldMaidGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDraw", mock.Anything).Return(domain.ErrGameEnded)

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
	// runCpuTurns and ArrangeTargetForHumanDraw not called when err != nil
	gameMock.AssertNotCalled(t, "CpuDraw")
	gameMock.AssertNotCalled(t, "ArrangeTargetForHumanDraw")
}

func TestOldMaidInteractor_Shuffle(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	t.Run("success calls ShuffleHumanHand", func(t *testing.T) {
		gameMock := new(interfaces.MockOldMaidGame)
		gameMock.On("ShuffleHumanHand").Return(nil)
		oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
		result := oi.Shuffle()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ShuffleHumanHand")
	})

	t.Run("error propagates to Output", func(t *testing.T) {
		gameMock := new(interfaces.MockOldMaidGame)
		gameMock.On("ShuffleHumanHand").Return(domain.ErrGameEnded)
		oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
		result := oi.Shuffle()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ShuffleHumanHand")
	})
}

func TestOldMaidInteractor_Reorder(t *testing.T) {
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	t.Run("success calls ReorderHumanHand", func(t *testing.T) {
		gameMock := new(interfaces.MockOldMaidGame)
		gameMock.On("ReorderHumanHand", []int{2, 0, 1}).Return(nil)
		oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
		result := oi.Reorder([]int{2, 0, 1})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ReorderHumanHand", []int{2, 0, 1})
	})

	t.Run("error propagates to Output", func(t *testing.T) {
		gameMock := new(interfaces.MockOldMaidGame)
		gameMock.On("ReorderHumanHand", []int{0, 0}).Return(domain.ErrInvalidIndices)
		oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
		result := oi.Reorder([]int{0, 0})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "ReorderHumanHand", []int{0, 0})
	})
}

func TestOldMaidInteractor_Draw_GameEndsAfterCpuTurns(t *testing.T) {
	// Game ends during runCpuTurns → ArrangeTargetForHumanDraw called but game ended
	mockOutput := `{"players":[]}`
	ompMock := new(presenter.MockOldMaidPresenter)
	ompMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockOldMaidGame)
	// First GetGameEndFlag call (Draw's first check): false
	// After PlayerDraw: GetGameEndFlag=false (allows runCpuTurns)
	// runCpuTurns loop: GetGameEndFlag=false, IsHumanTurn=false → CpuDraw
	// Second loop: GetGameEndFlag=true → exits loop
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()
	gameMock.On("PlayerDraw", mock.Anything).Return(nil)
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuDraw").Return(nil)
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("ArrangeTargetForHumanDraw").Return()

	oi := usecase.NewOldMaidInteractor(gameMock, ompMock)
	result := oi.Draw(0)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "ArrangeTargetForHumanDraw")
}
