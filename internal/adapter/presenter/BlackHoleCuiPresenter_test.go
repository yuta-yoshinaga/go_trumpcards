//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// bhState builds a BlackHole in an arbitrary state via JSON restore.
func bhState(t *testing.T, js string) *domain.BlackHole {
	t.Helper()
	g := domain.NewDefaultBlackHole()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestBlackHoleCuiPresenter_Output(t *testing.T) {
	p := new(presenter.BlackHoleCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Black Hole")
		assert.Contains(t, out, "ブラックホール")
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(bhState(t, `{"ph":1,"mc":51}`), nil), "ゲームクリア")
	})
	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(bhState(t, `{"ph":2}`), nil), "ゲームオーバー")
	})
	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(bhState(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		// Black hole top = 5, a fan top = 6 -> playable hint.
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":6,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(bhState(t, js)), "HINT")
		assert.NotEmpty(t, p.HintOutput(bhState(t, `{"ph":2}`)))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
