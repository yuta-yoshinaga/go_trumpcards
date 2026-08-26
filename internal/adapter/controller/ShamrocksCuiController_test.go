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

func newLlCuiSH() *controller.ShamrocksCuiController {
	g := domain.NewDefaultShamrocks()
	g.Reset()
	li := usecase.NewShamrocksInteractor(g, new(presenter.ShamrocksCuiPresenter))
	return controller.NewShamrocksCuiController(li)
}

func TestShamrocksCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newLlCuiSH().Exec("q"))
	})
	t.Run("reset renders the board", func(t *testing.T) {
		assert.Contains(t, newLlCuiSH().Exec("r"), "Shamrocks")
	})
	t.Run("no-arg commands", func(t *testing.T) {
		c := newLlCuiSH()
		for _, cmd := range []string{"rd", "ac", "u", "h", "l", "m 0 1", "m 0 f"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})
	t.Run("invalid move args", func(t *testing.T) {
		c := newLlCuiSH()
		assert.True(t, msgRejected(c.Exec("m")))
		assert.Contains(t, c.Exec("m x 1"), msgStem("invalidSourceFanDot"))
		assert.Contains(t, c.Exec("m 0 x"), msgStem("invalidDestinationFanOrF"))
	})
	t.Run("giveup ends the game", func(t *testing.T) {
		assert.NotEmpty(t, newLlCuiSH().Exec("g"))
	})
	t.Run("unknown command", func(t *testing.T) {
		assert.NotEmpty(t, newLlCuiSH().Exec("zzz"))
	})
}
