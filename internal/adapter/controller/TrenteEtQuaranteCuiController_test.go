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

func TestTrenteEtQuaranteCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockTrenteEtQuaranteInteractor {
		m := new(mockUsecases.MockTrenteEtQuaranteInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Bet", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultTrenteEtQuaranteConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTrenteEtQuaranteCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrenteEtQuaranteCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("bet", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrenteEtQuaranteCuiController(m).Exec("b 1 100"))
		m.AssertCalled(t, "Bet", domain.TrenteEtQuaranteBetRouge, 100)
	})
	t.Run("bet missing args", func(t *testing.T) {
		out := controller.NewTrenteEtQuaranteCuiController(newMock()).Exec("b")
		assert.True(t, msgRejected(out))
	})
	t.Run("bet invalid type", func(t *testing.T) {
		out := controller.NewTrenteEtQuaranteCuiController(newMock()).Exec("b 9 100")
		assert.Contains(t, out, msgStem("invalidBetTypeNoir"))
	})
	t.Run("bet invalid stake", func(t *testing.T) {
		out := controller.NewTrenteEtQuaranteCuiController(newMock()).Exec("b 0 xyz")
		assert.Contains(t, out, msgStem("invalidStakeDot"))
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrenteEtQuaranteCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set default bet", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTrenteEtQuaranteCuiController(m).Exec("sb 3"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewTrenteEtQuaranteCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewTrenteEtQuaranteCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewTrenteEtQuaranteCuiController(newMock()).Exec("zzz"))
	})
}
