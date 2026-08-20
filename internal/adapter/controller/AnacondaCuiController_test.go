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

func TestAnacondaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockAnacondaInteractor {
		m := new(mockUsecases.MockAnacondaInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Pass", mock.Anything).Return(mockOutput)
		m.On("Keep", mock.Anything).Return(mockOutput)
		m.On("Call").Return(mockOutput)
		m.On("Raise").Return(mockOutput)
		m.On("Fold").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultAnacondaConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewAnacondaCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("pass", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("p 0 1 2"))
		m.AssertCalled(t, "Pass", []int{0, 1, 2})
	})
	t.Run("keep", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("k 0 1 2 3 4"))
		m.AssertCalled(t, "Keep", []int{0, 1, 2, 3, 4})
	})
	t.Run("call", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("c"))
		m.AssertCalled(t, "Call")
	})
	t.Run("raise", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("ra"))
		m.AssertCalled(t, "Raise")
	})
	t.Run("fold", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("f"))
		m.AssertCalled(t, "Fold")
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set players", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("sp 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set players missing arg", func(t *testing.T) {
		out := controller.NewAnacondaCuiController(newMock()).Exec("sp")
		assert.Contains(t, out, msgStem("playerCountRequiredEGSp4"))
	})
	t.Run("set players invalid", func(t *testing.T) {
		out := controller.NewAnacondaCuiController(newMock()).Exec("sp 99")
		assert.Contains(t, out, msgStem("invalidPlayerCount37"))
	})
	t.Run("set ante", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("sa 20"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set chips", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("sc 500"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set rounds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewAnacondaCuiController(m).Exec("st 20"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewAnacondaCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewAnacondaCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewAnacondaCuiController(newMock()).Exec("zzz"))
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewAnacondaCuiController(m)
		assert.Contains(t, c.Exec("p 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
		assert.Contains(t, c.Exec("k 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
