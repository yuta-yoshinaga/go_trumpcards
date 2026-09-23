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

func TestBatakCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBatakInteractor {
		m := new(mockUsecases.MockBatakInteractor)
		m.On("GetConfig").Return(domain.DefaultBatakConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewBatakCuiController(newMock()).Exec("q"))
	})
	t.Run("quit quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewBatakCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBatakConfig())
	})

	t.Run("bid valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("b 5"))
		m.AssertCalled(t, "Bid", 5)
	})
	t.Run("bid pass with 0", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("b 0"))
		m.AssertCalled(t, "Bid", 0)
	})
	t.Run("pass command", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Bid", domain.BatakPassBid)
	})
	t.Run("bid long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("bid 5"))
		m.AssertCalled(t, "Bid", 5)
	})
	t.Run("bid no args", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("b"), msgStem("bidValueRequiredPass513"))
	})
	t.Run("bid invalid arg", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("b abc"), msgStem("invalidBidValue"))
	})
	t.Run("bid below min", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("b -1"), msgKey("invalidBidValue", "val", "-1"))
	})
	t.Run("bid above max", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("b 14"), msgKey("invalidBidValue", "val", "14"))
	})

	t.Run("play valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})
	t.Run("play long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("play 5"))
		m.AssertCalled(t, "Play", 5)
	})
	t.Run("play no args", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("p"), msgCardIndexRequired())
	})
	t.Run("play invalid arg", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("next n", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("sd 2"))
		expected := domain.DefaultBatakConfig()
		expected.CpuDifficulty = domain.BatakCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setdifficulty invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("sd 5"), msgInvalidCpuDifficultyPrefix())
	})
	t.Run("setdifficulty no args", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("sd"), msgCpuDifficultyRequired())
	})

	t.Run("setrounds valid", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("sr 7"))
		expected := domain.DefaultBatakConfig()
		expected.MaxRounds = 7
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setrounds zero is invalid", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("sr 0"), msgStem("invalidRoundCount1OrMore"))
	})
	t.Run("setrounds long form", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("setrounds 10"))
		expected := domain.DefaultBatakConfig()
		expected.MaxRounds = 10
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setrounds no args", func(t *testing.T) {
		assert.True(t, msgRejected(controller.NewBatakCuiController(newMock()).Exec("sr")))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewBatakCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("pass typo suggest", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("pas"), "pass")
	})
	t.Run("unknown command", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec("unknown"), "コマンドが不明です")
	})
	t.Run("empty command", func(t *testing.T) {
		assert.Contains(t, controller.NewBatakCuiController(newMock()).Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
