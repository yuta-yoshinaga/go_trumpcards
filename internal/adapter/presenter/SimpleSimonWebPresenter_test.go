//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSimpleSimonWebPresenter_Output(t *testing.T) {
	p := new(presenter.SimpleSimonWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"columns"`, `"completedSuits"`, `"canUndo"`, `"messageCode":"simplesimon.playing"`} {
			assert.Contains(t, out, frag)
		}
		assert.Contains(t, out, `"runStarts":[`)
		assert.Contains(t, out, `"legalTargets":{`)
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(ssState(t, `{"ph":1,"mc":5}`), nil), "simplesimon.gameClear")
		assert.Contains(t, p.Output(ssState(t, `{"ph":2}`), nil), "simplesimon.gameOver")
	})

	// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	t.Run("output carries the hint", func(t *testing.T) {
		// 9 の上に 8 を重ねられる並び。同スートなので合法手になる。
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.Output(ssState(t, js), nil), `"hint"`,
			"Output must carry the hint -- the frontend reads state.hint")
		assert.NotContains(t, p.Output(ssState(t, `{"ph":2}`), nil), `"hint"`)
	})

	t.Run("hint output", func(t *testing.T) {
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(ssState(t, js)), "simplesimon.hintAvailable")
		assert.Contains(t, p.HintOutput(ssState(t, `{"ph":2}`)), "simplesimon.noHint")
		assert.Contains(t, p.HintOutput(ssState(t, `{"ph":2}`)), `"runStarts":[[],[],[],[],[],[],[],[],[],[]]`)
		assert.Contains(t, p.HintOutput(ssState(t, `{"ph":2}`)), `"legalTargets":{}`)
	})

	t.Run("targets include ordered tableau and empty foundation arrays", func(t *testing.T) {
		g := ssState(t, `{"co":[[{"d":1,"v":9,"w":true},{"d":1,"v":7,"w":true}],[{"d":1,"v":8,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}],[{"d":2,"v":2,"w":true}]],"ph":0}`)
		var output map[string]any
		assert.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &output))
		assert.Equal(t, map[string]any{"tableau": []any{1.0}, "foundation": []any{}}, output["legalTargets"].(map[string]any)["tableau-0-1"])
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
