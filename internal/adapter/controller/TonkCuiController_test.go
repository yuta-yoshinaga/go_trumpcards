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

func TestTonkCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTonkInteractor {
		m := new(mockUsecases.MockTonkInteractor)
		m.On("GetConfig").Return(domain.DefaultTonkConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Knock", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTonkConfig())
	})

	t.Run("drawstock", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		assert.Equal(t, mockOutput, c.Exec("drawstock"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd"))
		assert.Equal(t, mockOutput, c.Exec("drawdiscard"))
		m.AssertCalled(t, "DrawFromDiscard")
	})

	t.Run("discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		assert.Equal(t, mockOutput, c.Exec("discard 5"))
		m.AssertCalled(t, "Discard", 3)
		m.AssertCalled(t, "Discard", 5)
	})

	t.Run("discard no args", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("d abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("knock with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("k 2"))
		assert.Equal(t, mockOutput, c.Exec("knock 7"))
		m.AssertCalled(t, "Knock", 2)
		m.AssertCalled(t, "Knock", 7)
	})

	t.Run("knock no args", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec("k"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("k abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultTonkConfig()
		expected.CpuDifficulty = domain.TonkCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty errors", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
		assert.Equal(t, msgInvalidCpuDifficulty("-1"), c.Exec("sd -1"))
		assert.Equal(t, msgInvalidCpuDifficulty("3"), c.Exec("sd 3"))
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 100"))
		expected := domain.DefaultTonkConfig()
		expected.PointLimit = 100
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit errors", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec("sl"), msgPointLimitRequired())
		assert.Contains(t, c.Exec("sl abc"), msgInvalidPointLimitPrefix())
		assert.Equal(t, msgInvalidPointLimit("0"), c.Exec("sl 0"))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTonkCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewTonkCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
