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

func TestBoliviaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockBoliviaInteractor {
		m := new(mockUsecases.MockBoliviaInteractor)
		m.On("GetConfig").Return(domain.DefaultBoliviaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("SkipMeld").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("GoOut").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBoliviaConfig())
	})

	t.Run("drawstock", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		assert.Equal(t, mockOutput, c.Exec("drawstock"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd 0,1"))
		m.AssertCalled(t, "DrawFromDiscard", []int{0, 1})
	})

	t.Run("meld with groups", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 0,1,2;3,4,5"))
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}, {3, 4, 5}})
	})

	t.Run("meld no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m"))
		m.AssertCalled(t, "Meld", ([][]int)(nil))
	})

	t.Run("skipmeld", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sm"))
		assert.Equal(t, mockOutput, c.Exec("skipmeld"))
		m.AssertCalled(t, "SkipMeld")
	})

	t.Run("discard with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard no args", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec("d"), msgCardIndexRequired())
	})

	t.Run("discard invalid arg", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec("d abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("goout", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("go"))
		assert.Equal(t, mockOutput, c.Exec("goout"))
		m.AssertCalled(t, "GoOut")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultBoliviaConfig()
		expected.CpuDifficulty = domain.BoliviaCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec("sd abc"), msgInvalidCpuDifficultyPrefix())
		assert.Contains(t, c.Exec("sd -1"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 8000"))
		expected := domain.DefaultBoliviaConfig()
		expected.PointLimit = 8000
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setlimit invalid", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec("sl abc"), msgInvalidPointLimitPrefix())
		assert.Contains(t, c.Exec("sl 0"), msgInvalidPointLimitPrefix())
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBoliviaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewBoliviaCuiController(newMock())
		assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	})
}
