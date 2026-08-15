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

func TestOmbreCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockOmbreInteractor {
		m := new(mockUsecases.MockOmbreInteractor)
		m.On("GetConfig").Return(domain.DefaultOmbreConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewOmbreCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewOmbreCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultOmbreConfig())
	})

	t.Run("bid pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid pass"))
		m.AssertCalled(t, "Bid", domain.OmbreBidNone, -1)
	})

	t.Run("bid entrar with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid entrar h"))
		m.AssertCalled(t, "Bid", domain.OmbreBidEntrar, domain.CardDesignHeart)
	})

	t.Run("bid solo with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid solo spades"))
		m.AssertCalled(t, "Bid", domain.OmbreBidSolo, domain.CardDesignSpade)
	})

	t.Run("bid entrar missing suit", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("bid entrar")
		assert.Contains(t, result, "Trump suit is required")
	})

	t.Run("bid entrar invalid suit", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("bid entrar zzz")
		assert.Contains(t, result, "Invalid trump suit")
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, "Bid action is required")
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, "Invalid bid action")
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultOmbreConfig()
		expected.CpuDifficulty = domain.OmbreCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewOmbreCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewOmbreCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
