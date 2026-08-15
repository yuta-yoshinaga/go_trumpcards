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

func TestKoenigrufenCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockKoenigrufenInteractor {
		m := new(mockUsecases.MockKoenigrufenInteractor)
		m.On("GetConfig").Return(domain.DefaultKoenigrufenConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("CallKing", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewKoenigrufenCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKoenigrufenConfig())
	})

	t.Run("bid rufer", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid rufer"))
		m.AssertCalled(t, "Bid", domain.KoenigrufenBidRufer)
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, "Bid is required")
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, "Invalid bid")
	})

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Pass")
	})

	t.Run("callking", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("callking 3"))
		m.AssertCalled(t, "CallKing", 3)
	})

	t.Run("callking no args", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("callking")
		assert.Contains(t, result, "King suit is required")
	})

	t.Run("callking invalid", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("callking 9")
		assert.Contains(t, result, "Invalid suit")
	})

	t.Run("discard six cards", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 0 1 2 3 4 5"))
		m.AssertCalled(t, "Discard", []int{0, 1, 2, 3, 4, 5})
	})

	t.Run("discard too few", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("discard 0 1")
		assert.Contains(t, result, "Six card indices are required")
	})

	t.Run("discard invalid index", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("discard 0 1 2 3 4 x")
		assert.Contains(t, result, msgInvalidCardIndexPrefix())
	})

	t.Run("play card", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 3"))
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("play no args", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("play")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("next / nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextTrick")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultKoenigrufenConfig()
		expected.CpuDifficulty = domain.KoenigrufenCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("hint / log", func(t *testing.T) {
		m := newMock()
		c := controller.NewKoenigrufenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "Hint")
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		result := controller.NewKoenigrufenCuiController(newMock()).Exec("zzz")
		assert.Contains(t, result, "コマンドが不明です")
	})
}
