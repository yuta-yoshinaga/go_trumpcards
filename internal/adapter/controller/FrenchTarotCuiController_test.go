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

func TestFrenchTarotCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockFrenchTarotInteractor {
		m := new(mockUsecases.MockFrenchTarotInteractor)
		m.On("GetConfig").Return(domain.DefaultFrenchTarotConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewFrenchTarotCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultFrenchTarotConfig())
	})

	t.Run("bid petite", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid petite"))
		m.AssertCalled(t, "Bid", domain.FrenchTarotBidPetite)
	})

	t.Run("bid gardecontre", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid gardecontre"))
		m.AssertCalled(t, "Bid", domain.FrenchTarotBidGardeContre)
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, "Bid is required")
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, "Invalid bid")
	})

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Pass")
	})

	t.Run("discard six cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 0 1 2 3 4 5"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2, 3, 4, 5})
	})

	t.Run("discard too few", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("discard 0 1")
		assert.Contains(t, result, "Six card indices are required")
	})

	t.Run("discard invalid index", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("discard 0 1 2 3 4 x")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultFrenchTarotConfig()
		expected.CpuDifficulty = domain.FrenchTarotCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewFrenchTarotCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewFrenchTarotCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
