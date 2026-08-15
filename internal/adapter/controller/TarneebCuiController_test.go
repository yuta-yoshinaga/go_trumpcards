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

func TestTarneebCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTarneebInteractor {
		m := new(mockUsecases.MockTarneebInteractor)
		m.On("GetConfig").Return(domain.DefaultTarneebConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("DeclareTrump", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTarneebCuiController(newMock()).Exec("q"))
	})

	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTarneebConfig())
	})

	t.Run("bid pass", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("b 0"))
		m.AssertCalled(t, "Bid", 0)
	})
	t.Run("bid valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("bid 9"))
		m.AssertCalled(t, "Bid", 9)
	})
	t.Run("bid no args", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec("b"), msgStem("bidValueRequiredPass713"))
	})
	t.Run("bid above max", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec("b 14"), msgStem("invalidBidValue"))
	})

	t.Run("trump valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("t 1"))
		m.AssertCalled(t, "DeclareTrump", domain.CardDesignSpade)
	})
	t.Run("trump long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("trump 4"))
		m.AssertCalled(t, "DeclareTrump", domain.CardDesignDiamond)
	})
	t.Run("trump no args", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec("t"), msgStem("trumpSuitRequiredNames"))
	})
	t.Run("trump invalid suit", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec("t 9"), msgStem("invalidSuit"))
	})

	t.Run("play valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("sd 2"))
		expected := domain.DefaultTarneebConfig()
		expected.CpuDifficulty = domain.TarneebCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("sl 41"))
		expected := domain.DefaultTarneebConfig()
		expected.PointLimit = 41
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setminbid valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("sm 8"))
		expected := domain.DefaultTarneebConfig()
		expected.MinBid = 8
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTarneebCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec("unknown"), "コマンドが不明です")
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Contains(t, controller.NewTarneebCuiController(newMock()).Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
