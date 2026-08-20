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

func newDkCui() *controller.DoubleKlondikeCuiController {
	g := domain.NewDefaultDoubleKlondike()
	g.Reset()
	di := usecase.NewDoubleKlondikeInteractor(g, new(presenter.DoubleKlondikeCuiPresenter))
	return controller.NewDoubleKlondikeCuiController(di)
}

func TestDoubleKlondikeCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newDkCui().Exec("q"))
	})
	t.Run("reset renders the board", func(t *testing.T) {
		assert.Contains(t, newDkCui().Exec("r"), "Double Klondike")
	})
	t.Run("commands produce output", func(t *testing.T) {
		c := newDkCui()
		for _, cmd := range []string{"d", "mwf", "mwt 0", "mtf 0", "mtt 0 0 1", "ac", "u", "h", "l"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})
	t.Run("invalid move args", func(t *testing.T) {
		c := newDkCui()
		assert.True(t, msgRejected(c.Exec("mwt")))
		assert.True(t, msgRejected(c.Exec("mtt 0 0")))
		assert.Equal(t, msgKey("invalidArgumentNotInteger", "val", "x"), c.Exec("mwt x"))
	})
	t.Run("giveup + unknown", func(t *testing.T) {
		assert.NotEmpty(t, newDkCui().Exec("g"))
		assert.NotEmpty(t, newDkCui().Exec("zzz"))
	})
}
