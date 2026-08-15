//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newBeggarMyNeighbourCuiMock() *mockUsecases.MockBeggarMyNeighbourInteractor {
	m := new(mockUsecases.MockBeggarMyNeighbourInteractor)
	m.On("Reset").Return("reset-ok")
	m.On("ResetWithConfig", mock.Anything).Return("reset-ok")
	m.On("Step").Return("step-ok")
	m.On("AutoPlay").Return("autoplay-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultBeggarMyNeighbourConfig())
	return m
}

func TestBeggarMyNeighbourCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		c := controller.NewBeggarMyNeighbourCuiController(newBeggarMyNeighbourCuiMock())
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newBeggarMyNeighbourCuiMock()
		assert.Equal(t, "reset-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec("r"))
	})

	for _, cmd := range []string{"s", "step"} {
		t.Run(cmd, func(t *testing.T) {
			m := newBeggarMyNeighbourCuiMock()
			assert.Equal(t, "step-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec(cmd))
			m.AssertCalled(t, "Step")
		})
	}

	for _, cmd := range []string{"a", "autoplay"} {
		t.Run(cmd, func(t *testing.T) {
			m := newBeggarMyNeighbourCuiMock()
			assert.Equal(t, "autoplay-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec(cmd))
			m.AssertCalled(t, "AutoPlay")
		})
	}

	for _, cmd := range []string{"l", "log"} {
		t.Run(cmd, func(t *testing.T) {
			m := newBeggarMyNeighbourCuiMock()
			assert.Equal(t, "log-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec(cmd))
		})
	}
}

// #5390: `sm abc` は最大ラウンド数を既定値で上書きし直していた。設定が変わった
// ようにも見えず、局だけが静かに配り直される。
func TestBeggarMyNeighbourCuiController_SetMax(t *testing.T) {
	t.Run("a value is applied", func(t *testing.T) {
		m := newBeggarMyNeighbourCuiMock()
		assert.Equal(t, "reset-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec("sm 42"))
		cfg := domain.DefaultBeggarMyNeighbourConfig()
		cfg.MaxRounds = 42
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("a typo is refused", func(t *testing.T) {
		m := newBeggarMyNeighbourCuiMock()
		out := controller.NewBeggarMyNeighbourCuiController(m).Exec("sm abc")
		assert.Equal(t, msgKey("invalidMaxRounds", "val", "abc"), out)
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	// **省略形も一緒に見る。** 拒否側だけ直して既定値を消してもテストが緑のまま
	// になるので、既定値が残っていることを同じ場所で押さえる。
	t.Run("no argument keeps the default", func(t *testing.T) {
		m := newBeggarMyNeighbourCuiMock()
		assert.Equal(t, "reset-ok", controller.NewBeggarMyNeighbourCuiController(m).Exec("sm"))
		cfg := domain.DefaultBeggarMyNeighbourConfig()
		cfg.MaxRounds = domain.BeggarMyNeighbourDefaultMaxRounds
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})
}
