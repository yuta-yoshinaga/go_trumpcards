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

// Big Two had no controller test at all -- and no interactor mock to write one
// with -- which is why its rejection branch showed as 0% covered.
func TestBigTwoCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockBigTwoInteractor {
		m := new(mockUsecases.MockBigTwoInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultBigTwoConfig())
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log output")
		return m
	}

	t.Run("play with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewBigTwoCuiController(m)

		assert.Equal(t, mockOutput, c.Exec("p 0 1"))

		m.AssertCalled(t, "Play", []int{0, 1})
	})

	// **引数なしがパス。** `p [idx ...]` の規則そのものなので、空スライスで
	// 呼ばれることを固定しておく。
	t.Run("play with no index passes", func(t *testing.T) {
		m := newMock()
		c := controller.NewBigTwoCuiController(m)

		assert.Equal(t, mockOutput, c.Exec("p"))

		m.AssertCalled(t, "Play", []int{})
	})

	// **落として残りで実行しない。** `p 0 zz` で zz を捨てると、ペアのつもりが
	// 単札という別の合法手になる (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBigTwoCuiController(m)

		assert.Contains(t, c.Exec("p 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")

		m.AssertNotCalled(t, "Play", mock.Anything)
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewBigTwoCuiController(m)

		assert.Equal(t, mockOutput, c.Exec("sd 1"))

		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("action log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBigTwoCuiController(m)

		assert.Equal(t, "log output", c.Exec("l"))
	})
}
