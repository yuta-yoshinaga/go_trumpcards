//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDoubleKlondikeWebPresenter_Output(t *testing.T) {
	p := new(presenter.DoubleKlondikeWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"tableau"`, `"stockCount"`, `"waste"`, `"foundation"`, `"isStalemate"`, `"faceUp"`, `"messageCode":"doubleklondike.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(dkState(t, `{"ph":1,"mc":9}`), nil), "doubleklondike.gameClear")
		assert.Contains(t, p.Output(dkState(t, `{"ph":2}`), nil), "doubleklondike.gameOver")
	})

	// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	t.Run("output carries the hint", func(t *testing.T) {
		// 廃札にスペードのエースを表向きで置く。台札へ動かせるのでヒントが必ず出る。
		js := `{"wa":[{"d":1,"v":1,"w":true}],"ph":0}`
		assert.Contains(t, p.Output(dkState(t, js), nil), `"hint"`,
			"Output must carry the hint -- the frontend reads state.hint")
		assert.NotContains(t, p.Output(dkState(t, `{"ph":2}`), nil), `"hint"`)
	})

	t.Run("hint output", func(t *testing.T) {
		js := `{"wa":[{"d":1,"v":1,"w":true}],"ph":0}`
		assert.Contains(t, p.HintOutput(dkState(t, js)), "doubleklondike.hintAvailable")
		assert.Contains(t, p.HintOutput(dkState(t, `{"ph":2}`)), "doubleklondike.noHint")
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultDoubleKlondike()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
