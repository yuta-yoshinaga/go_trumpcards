//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCurdsAndWheyWebPresenter_Output(t *testing.T) {
	p := new(presenter.CurdsAndWheyWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"columns"`, `"completedSuits"`, `"canUndo"`, `"messageCode":"curdsandwhey.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(ssStateCW(t, `{"ph":1,"mc":5}`), nil), "curdsandwhey.gameClear")
		assert.Contains(t, p.Output(ssStateCW(t, `{"ph":2}`), nil), "curdsandwhey.gameOver")
	})

	// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	t.Run("output carries the hint", func(t *testing.T) {
		// 9 の上に 8 を重ねられる並び。同スートなので合法手になる。
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.Output(ssStateCW(t, js), nil), `"hint"`,
			"Output must carry the hint -- the frontend reads state.hint")
		assert.NotContains(t, p.Output(ssStateCW(t, `{"ph":2}`), nil), `"hint"`)
	})

	t.Run("hint output", func(t *testing.T) {
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(ssStateCW(t, js)), "curdsandwhey.hintAvailable")
		assert.Contains(t, p.HintOutput(ssStateCW(t, `{"ph":2}`)), "curdsandwhey.noHint")
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
