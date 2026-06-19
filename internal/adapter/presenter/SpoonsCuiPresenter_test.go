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

	// i18n resolves to the ja translations in this build, so assert on the
	// rendered Japanese text that uniquely identifies each phase.
	t.Run("initial pass state", func(t *testing.T) {
		g := setupSpoonsTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Spoons")
		assert.Contains(t, result, "ラウンド")
		assert.Contains(t, result, "あなたの番です")
	})

	t.Run("error", func(t *testing.T) {
		g := setupSpoonsTest()
		assert.Contains(t, p.Output(g, errors.New("oops")), "oops")
	})

	t.Run("grab prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseGrab, "gw": true})
		assert.Contains(t, p.Output(g, nil), "g (grab)")
	})

	t.Run("round end with loser", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ph": domain.SpoonsPhaseRoundEnd, "rl": 2})
		out := p.Output(g, nil)
		assert.Contains(t, out, "n (next)")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ci": 1})
		assert.Contains(t, p.Output(g, nil), "CPU")
	})

	t.Run("human win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 0, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "あなたの勝ちです")
	})

	t.Run("cpu win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 1, "ph": domain.SpoonsPhaseGameEnd})
		assert.Contains(t, p.Output(g, nil), "の勝ちです")
	})

	t.Run("eliminated and has-spoon shown", func(t *testing.T) {
		g := setupSpoonsTest()
		g.GetPlayer(3).SetEliminated(true)
		g.GetPlayer(1).SetHasSpoon(true)
		assert.Contains(t, p.Output(g, nil), "脱落")
	})
}

func TestSpoonsCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoonsCuiPresenter)
	g := setupSpoonsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
