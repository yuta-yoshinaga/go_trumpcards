//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSpoonsCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.SpoonsCuiPresenter)

	// NOTE: i18n is not initialised in this test build, so i18n.T returns the
	// raw "<file>.<key>" namespace keys verbatim (same convention the other CUI
	// presenter tests rely on). Assert on those keys, which uniquely identify
	// each rendered phase.
	t.Run("initial pass state", func(t *testing.T) {
		g := setupSpoonsTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "spoons.helpTitle")
		assert.Contains(t, result, "spoons.promptPass")
	})

	t.Run("error", func(t *testing.T) {
		g := setupSpoonsTest()
		assert.Contains(t, p.Output(g, errors.New("oops")), "oops")
	})

	t.Run("grab prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseGrab, "gw": true})
		assert.Contains(t, p.Output(g, nil), "spoons.promptGrab")
	})

	t.Run("round end with loser", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseRoundEnd, "rl": 2})
		out := p.Output(g, nil)
		assert.Contains(t, out, "spoons.promptNextRound")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ci": 1})
		assert.Contains(t, p.Output(g, nil), "spoons.promptCpuTurn")
	})

	t.Run("human win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 0, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "spoons.winHuman")
	})

	t.Run("cpu win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 1, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "spoons.winCpu")
	})

	t.Run("eliminated and has-spoon shown", func(t *testing.T) {
		g := setupSpoonsTest()
		g.GetPlayer(3).SetEliminated(true)
		g.GetPlayer(1).SetHasSpoon(true)
		assert.Contains(t, p.Output(g, nil), "spoons.playerEliminated")
	})
}

func TestSpoonsCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoonsCuiPresenter)
	g := setupSpoonsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
