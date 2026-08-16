//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// dkState builds a DoubleKlondike in an arbitrary state via JSON restore.
func dkState(t *testing.T, js string) *domain.DoubleKlondike {
	t.Helper()
	g := domain.NewDefaultDoubleKlondike()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestDoubleKlondikeCuiPresenter_Output(t *testing.T) {
	p := new(presenter.DoubleKlondikeCuiPresenter)

	t.Run("playing board (with face-down cards)", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Double Klondike")
		assert.Contains(t, out, "ストック")
		assert.Contains(t, out, "##") // face-down marker
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(dkState(t, `{"ph":1,"mc":120}`), nil), "ゲームクリア")
	})
	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(dkState(t, `{"ph":2}`), nil), "ゲームオーバー")
	})
	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(dkState(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		// Waste holds an Ace -> foundation hint.
		js := `{"wa":[{"d":1,"v":1,"w":true}],"ph":0}`
		out := p.HintOutput(dkState(t, js))
		assert.Contains(t, out, "HINT")
		// Zone identifiers are localised (ja), not raw "waste"/"foundation".
		assert.Contains(t, out, "ウェイスト")
		assert.Contains(t, out, "ファウンデーション")
		assert.NotContains(t, out, "waste")
		assert.NotContains(t, out, "foundation")
		assert.NotEmpty(t, p.HintOutput(dkState(t, `{"ph":2}`)))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}

// **押せない操作をそもそも見せない、が Web 側の設計。** CUI には可否が出ておらず、
// `u` を打って初めてエラーで分かる作りだった (#5680)。CanUndo はインタフェースに
// 在るのに、どの CUI presenter も読んでいなかった。
func TestDoubleKlondikeCuiPresenter_ShowsWhetherUndoIsAvailable(t *testing.T) {
	p := new(presenter.DoubleKlondikeCuiPresenter)

	t.Run("a fresh deal has nothing to undo", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		require.False(t, g.CanUndo(), "配った直後は戻せないはず (前提が崩れたらテストの意味が消える)")

		out := p.Output(g, nil)
		assert.Contains(t, out, i18n.T("cuiSolitaireUndoUnavailable"))
		assert.NotContains(t, out, i18n.T("cuiSolitaireUndoAvailable"))
	})

	t.Run("after a move it says the move can be undone", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		// ストックをめくれば必ず 1 手進む。どの配りでも打てる唯一の手。
		require.NoError(t, g.Draw())
		require.True(t, g.CanUndo())

		out := p.Output(g, nil)
		assert.Contains(t, out, i18n.T("cuiSolitaireUndoAvailable"))
		assert.NotContains(t, out, i18n.T("cuiSolitaireUndoUnavailable"))
	})
}
