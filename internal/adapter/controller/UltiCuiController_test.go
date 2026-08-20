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

func TestUltiCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockUltiInteractor {
		m := new(mockUsecases.MockUltiInteractor)
		m.On("GetConfig").Return(domain.DefaultUltiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewUltiCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewUltiCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultUltiConfig())
	})

	t.Run("bid party with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid party h"))
		m.AssertCalled(t, "Bid", domain.UltiContractParty, domain.CardDesignHeart)
	})

	t.Run("bid betli", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid betli"))
		m.AssertCalled(t, "Bid", domain.UltiContractBetli, -1)
	})

	t.Run("bid durchmarsch", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid durchmarsch"))
		m.AssertCalled(t, "Bid", domain.UltiContractDurchmarsch, -1)
	})

	t.Run("bid party missing suit", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("bid party")
		assert.Contains(t, result, msgStem("trumpSuitRequiredWords"))
	})

	t.Run("bid party invalid suit", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("bid party zzz")
		assert.Contains(t, result, msgStem("invalidTrumpSuitSCHD"))
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, msgStem("bidActionRequiredParty"))
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, msgStem("invalidBidActionParty"))
	})

	t.Run("discard two cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 0 5"))
		m.AssertCalled(t, "Discard", []int{0, 5})
	})

	t.Run("discard too few", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("discard 0")
		assert.Contains(t, result, msgStem("twoIndicesRequiredDiscard"))
	})

	t.Run("discard invalid index", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("discard 0 x")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultUltiConfig()
		expected.CpuDifficulty = domain.UltiCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewUltiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewUltiCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
