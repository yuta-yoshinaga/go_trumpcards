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

func TestKarnoffelCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockKarnoffelInteractor {
		m := new(mockUsecases.MockKarnoffelInteractor)
		m.On("GetConfig").Return(domain.DefaultKarnoffelConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("PlayCard", mock.Anything).Return(mockOutput)
		m.On("NextHand").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit and reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewKarnoffelCuiController(m)
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKarnoffelConfig())
	})

	// **手札は 5 枚。**issue の 12 枚なら 11 まで通ってしまう。
	t.Run("play and next", func(t *testing.T) {
		m := newMock()
		c := controller.NewKarnoffelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 4"))
		m.AssertCalled(t, "PlayCard", 4)
		assert.Equal(t, mockOutput, c.Exec("play 0"))
		m.AssertCalled(t, "PlayCard", 0)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextHand")
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p 5"), msgInvalidCardIndexPrefix())
		assert.Contains(t, c.Exec("p -1"), msgInvalidCardIndexPrefix())
	})

	t.Run("log and unknown", func(t *testing.T) {
		m := newMock()
		c := controller.NewKarnoffelCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
