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

func newLlCui() *controller.LaBelleLucieCuiController {
	g := domain.NewDefaultLaBelleLucie()
	g.Reset()
	li := usecase.NewLaBelleLucieInteractor(g, new(presenter.LaBelleLucieCuiPresenter))
	return controller.NewLaBelleLucieCuiController(li)
}

func TestLaBelleLucieCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newLlCui().Exec("q"))
	})
	t.Run("reset renders the board", func(t *testing.T) {
		assert.Contains(t, newLlCui().Exec("r"), "La Belle Lucie")
	})
	t.Run("no-arg commands", func(t *testing.T) {
		c := newLlCui()
		for _, cmd := range []string{"rd", "ac", "u", "h", "l", "m 0 1", "m 0 f"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})
	t.Run("invalid move args", func(t *testing.T) {
		c := newLlCui()
		assert.True(t, msgRejected(c.Exec("m")))
		assert.Contains(t, c.Exec("m x 1"), "Invalid source")
		assert.Contains(t, c.Exec("m 0 x"), "Invalid destination")
	})
	t.Run("giveup ends the game", func(t *testing.T) {
		assert.NotEmpty(t, newLlCui().Exec("g"))
	})
	t.Run("unknown command", func(t *testing.T) {
		assert.NotEmpty(t, newLlCui().Exec("zzz"))
	})
}
