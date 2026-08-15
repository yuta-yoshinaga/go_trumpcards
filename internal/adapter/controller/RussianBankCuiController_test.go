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

// newRbCui wires a real interactor (real domain + real CUI presenter) to the CUI controller.
func newRbCui() *controller.RussianBankCuiController {
	g := domain.NewDefaultRussianBank()
	g.Reset()
	ti := usecase.NewRussianBankInteractor(g, new(presenter.RussianBankCuiPresenter))
	return controller.NewRussianBankCuiController(ti)
}

func TestRussianBankCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", newRbCui().Exec("q"))
		assert.Equal(t, "bye.", newRbCui().Exec("quit"))
	})

	t.Run("reset renders the board", func(t *testing.T) {
		out := newRbCui().Exec("r")
		assert.Contains(t, out, "Russian Bank")
	})

	t.Run("move commands produce output", func(t *testing.T) {
		c := newRbCui()
		for _, cmd := range []string{"pf r", "pf w", "pf or", "pf ow", "pf t0", "mt r 0", "mt w 1", "d", "u", "h", "l"} {
			assert.NotEmpty(t, c.Exec(cmd), "command %q should produce output", cmd)
		}
	})

	t.Run("invalid arguments report errors", func(t *testing.T) {
		c := newRbCui()
		assert.True(t, msgRejected(c.Exec("pf")))
		assert.Contains(t, c.Exec("pf zzz"), "Invalid source")
		assert.Contains(t, c.Exec("mt r"), "Usage")
		assert.Contains(t, c.Exec("mt zzz 0"), "Invalid source")
		assert.Contains(t, c.Exec("mt r 9"), "Invalid column")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		c := newRbCui()
		assert.NotEmpty(t, c.Exec("sd 2"))
		assert.NotEmpty(t, c.Exec("sd")) // missing arg -> usage message
	})

	t.Run("stop reports no violation", func(t *testing.T) {
		// Right after a fresh deal it is the human turn with no CPU lapse to catch.
		assert.NotEmpty(t, newRbCui().Exec("s"))
	})

	t.Run("unknown command", func(t *testing.T) {
		assert.NotEmpty(t, newRbCui().Exec("zzz"))
	})
}
