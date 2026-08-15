package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDoudizhuCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockDoudizhuInteractor {
		m := new(mockUsecases.MockDoudizhuInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultDoudizhuConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(`{"entries":[]}`)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewDoudizhuCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset command r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play command p with no index (pass)", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})
	})

	t.Run("play command p with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1"))
		m.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("bid command b with value", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 2"))
		m.AssertCalled(t, "Bid", 2)
	})

	t.Run("bid command b with no value defaults to pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b"))
		m.AssertCalled(t, "Bid", 0)
	})

	t.Run("bid command rejects invalid value", func(t *testing.T) {
		c := controller.NewDoudizhuCuiController(newMock())
		assert.Contains(t, c.Exec("b 9"), msgStem("invalidBidValue03Pass"))
	})

	t.Run("set difficulty command sd", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		c.Exec("sd 2")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Equal(t, `{"entries":[]}`, c.Exec("log"))
	})

	// **落として残りで実行しない。** 打ち間違いを捨てると、プレイヤーが
	// 選んでいない組み合わせが実行される (issue #5390)。
	t.Run("refuses a mistyped index", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoudizhuCuiController(m)
		assert.Contains(t, c.Exec("p 0 zz"), msgInvalidCardIndexPrefix(),
			"a mistyped index must be refused, not dropped")
	})
}
