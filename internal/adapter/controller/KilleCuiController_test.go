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

func TestKilleCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockKilleInteractor {
		m := new(mockUsecases.MockKilleInteractor)
		m.On("GetConfig").Return(domain.DefaultKilleConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Exchange").Return(mockOutput)
		m.On("Satisfied").Return(mockOutput)
		m.On("Reenter").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewKilleCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewKilleCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKilleConfig())
	})

	t.Run("exchange and satisfied", func(t *testing.T) {
		m := newMock()
		c := controller.NewKilleCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("e"))
		assert.Equal(t, mockOutput, c.Exec("exchange"))
		assert.Equal(t, mockOutput, c.Exec("s"))
		assert.Equal(t, mockOutput, c.Exec("satisfied"))
		m.AssertCalled(t, "Exchange")
		m.AssertCalled(t, "Satisfied")
	})

	t.Run("reenter and nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewKilleCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("re"))
		assert.Equal(t, mockOutput, c.Exec("reenter"))
		assert.Equal(t, mockOutput, c.Exec("nr"))
		assert.Equal(t, mockOutput, c.Exec("nextround"))
		m.AssertCalled(t, "Reenter")
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setstake valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewKilleCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 7"))
		expected := domain.DefaultKilleConfig()
		expected.Stake = 7
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("setstake errors", func(t *testing.T) {
		c := controller.NewKilleCuiController(newMock())
		assert.Contains(t, c.Exec("st"), "required")
		assert.Contains(t, c.Exec("st abc"), "Invalid stake")
		assert.Contains(t, c.Exec("st 0"), "Invalid stake")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewKilleCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewKilleCuiController(newMock())
		assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
	})
}
