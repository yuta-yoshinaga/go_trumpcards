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

func TestGermanSoloCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockGermanSoloInteractor {
		m := new(mockUsecases.MockGermanSoloInteractor)
		m.On("GetConfig").Return(domain.DefaultGermanSoloConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("CallAce", mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewGermanSoloCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewGermanSoloCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultGermanSoloConfig())
	})

	t.Run("bid pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid pass"))
		m.AssertCalled(t, "Bid", domain.GermanSoloBidNone, -1)
	})

	t.Run("bid frage with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid frage h"))
		m.AssertCalled(t, "Bid", domain.GermanSoloBidFrage, domain.CardDesignHeart)
	})

	t.Run("bid solo with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid solo spades"))
		m.AssertCalled(t, "Bid", domain.GermanSoloBidSolo, domain.CardDesignSpade)
	})

	t.Run("bid tout with suit", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid tout d"))
		m.AssertCalled(t, "Bid", domain.GermanSoloBidTout, domain.CardDesignDiamond)
	})

	// **エース呼びフェーズを抜ける唯一の入力。** これを落とすと落札の直後で
	// 盤面が固まり、play は「フェーズが違う」で弾かれ続ける。
	t.Run("ace call", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ace c"))
		m.AssertCalled(t, "CallAce", domain.CardDesignClover)
		assert.Equal(t, mockOutput, controller.NewGermanSoloCuiController(newMock()).Exec("a hearts"))
	})

	t.Run("ace call rejects a bad suit", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("ace zzz")
		assert.Contains(t, result, msgStem("invalidAceSuitSCHD"))
		assert.Contains(t, controller.NewGermanSoloCuiController(newMock()).Exec("ace"), msgStem("aceSuitRequiredWords"))
	})

	t.Run("bid frage missing suit", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("bid frage")
		assert.Contains(t, result, msgStem("trumpSuitRequiredWords"))
	})

	t.Run("bid frage invalid suit", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("bid frage zzz")
		assert.Contains(t, result, msgStem("invalidTrumpSuitSCHD"))
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, msgStem("bidActionRequiredFrage"))
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, msgStem("invalidBidActionFrage"))
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultGermanSoloConfig()
		expected.CpuDifficulty = domain.GermanSoloCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewGermanSoloCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewGermanSoloCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
