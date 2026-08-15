//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCourtPieceCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCourtPieceInteractor {
		m := new(mockUsecases.MockCourtPieceInteractor)
		m.On("GetConfig").Return(domain.DefaultCourtPieceConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DeclareTrump", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCourtPieceCuiController(newMock()).Exec("q"))
	})

	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCourtPieceConfig())
	})

	t.Run("trump valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("t 1"))
		m.AssertCalled(t, "DeclareTrump", domain.CardDesignSpade)
	})
	t.Run("trump long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("trump 4"))
		m.AssertCalled(t, "DeclareTrump", domain.CardDesignDiamond)
	})
	t.Run("trump no args", func(t *testing.T) {
		assert.Contains(t, controller.NewCourtPieceCuiController(newMock()).Exec("t"), msgStem("trumpSuitRequiredNames"))
	})
	t.Run("trump invalid suit", func(t *testing.T) {
		assert.Contains(t, controller.NewCourtPieceCuiController(newMock()).Exec("t 9"), msgStem("invalidSuit"))
	})

	t.Run("play valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("sd 2"))
		expected := domain.DefaultCourtPieceConfig()
		expected.CpuDifficulty = domain.CourtPieceCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("sl 9"))
		expected := domain.DefaultCourtPieceConfig()
		expected.PointLimit = 9
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCourtPieceCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewCourtPieceCuiController(newMock()).Exec("unknown"), "コマンドが不明です")
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Contains(t, controller.NewCourtPieceCuiController(newMock()).Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
