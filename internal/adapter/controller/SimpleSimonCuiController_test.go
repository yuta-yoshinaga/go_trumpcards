//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newSsCui() *controller.SimpleSimonCuiController {
	g := domain.NewDefaultSimpleSimon()
	g.Reset()
	si := usecase.NewSimpleSimonInteractor(g, new(presenter.SimpleSimonCuiPresenter))
	return controller.NewSimpleSimonCuiController(si)
}

func TestSimpleSimonCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newSsCui().Exec("q"))
	})
	t.Run("reset renders the board", func(t *testing.T) {
		assert.Contains(t, newSsCui().Exec("r"), "Simple Simon")
	})
	t.Run("no-arg + move commands", func(t *testing.T) {
		c := newSsCui()
		for _, cmd := range []string{"u", "h", "l", "m 0 0 1"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})
	t.Run("invalid move args", func(t *testing.T) {
		c := newSsCui()
		assert.True(t, msgRejected(c.Exec("m 0 0")))
		assert.True(t, msgRejected(c.Exec("m x 0 1")))
	})
	t.Run("giveup + unknown", func(t *testing.T) {
		assert.NotEmpty(t, newSsCui().Exec("g"))
		assert.NotEmpty(t, newSsCui().Exec("zzz"))
	})
}
