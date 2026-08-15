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

func newBhCui() *controller.BlackHoleCuiController {
	g := domain.NewDefaultBlackHole()
	g.Reset()
	li := usecase.NewBlackHoleInteractor(g, new(presenter.BlackHoleCuiPresenter))
	return controller.NewBlackHoleCuiController(li)
}

func TestBlackHoleCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newBhCui().Exec("q"))
	})
	t.Run("reset renders the board", func(t *testing.T) {
		assert.Contains(t, newBhCui().Exec("r"), "Black Hole")
	})
	t.Run("no-arg commands", func(t *testing.T) {
		c := newBhCui()
		for _, cmd := range []string{"u", "h", "l", "m 0"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})
	t.Run("invalid move args", func(t *testing.T) {
		c := newBhCui()
		assert.True(t, msgRejected(c.Exec("m")))
		assert.Contains(t, c.Exec("m x"), msgStem("invalidFanIndexDot"))
	})
	t.Run("giveup ends the game", func(t *testing.T) {
		assert.NotEmpty(t, newBhCui().Exec("g"))
	})
	t.Run("unknown command", func(t *testing.T) {
		assert.NotEmpty(t, newBhCui().Exec("zzz"))
	})
}
