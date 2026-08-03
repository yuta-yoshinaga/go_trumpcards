//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBlackHoleWebPresenter_Output(t *testing.T) {
	p := new(presenter.BlackHoleWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"fans"`, `"blackHole"`, `"isStalemate"`, `"messageCode":"blackhole.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(bhState(t, `{"ph":1,"mc":51}`), nil), "blackhole.gameClear")
		assert.Contains(t, p.Output(bhState(t, `{"ph":2}`), nil), "blackhole.gameOver")
	})

	// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	// **「手詰まりならヒントは nil」をドメインの性質として固定する。**
	// Output のゲートから `!IsStalemate()` を外した根拠がこれで、崩れたら
	// ゲートを戻す必要がある (#4542 のレビュー指摘)。
	t.Run("a stalemate never has a hint", func(t *testing.T) {
		// 穴の一番上が 5、どの扇の一番上も ±1 でない → 合法手なし。
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":10,"w":true}]],"ph":0}`
		g := bhState(t, js)
		assert.True(t, g.IsStalemate(), "fixture must actually be a stalemate")
		assert.Nil(t, g.GetHint(), "a stalemate must not produce a hint")
		assert.NotContains(t, p.Output(g, nil), `"hint"`)
	})

	t.Run("output carries the hint", func(t *testing.T) {
		// 穴の一番上が 5、fan0 の一番上が 6（±1）なので fan 0 が勧められる。
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":6,"w":true}]],"ph":0}`
		assert.Contains(t, p.Output(bhState(t, js), nil), `"hint"`,
			"Output must carry the hint -- the frontend reads state.hint")
		assert.NotContains(t, p.Output(bhState(t, `{"ph":2}`), nil), `"hint"`)
	})

	t.Run("hint output carries recommended fan and full board", func(t *testing.T) {
		// Hole top 5, fan0 top 6 (±1) -> the domain recommends fan 0.
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":6,"w":true}]],"ph":0}`
		out := p.HintOutput(bhState(t, js))
		for _, frag := range []string{
			"blackhole.hintAvailable",
			`"hint":{"fan":0}`,
			`"fans":[[`, // the board is preserved so the tableau does not blank out
			`"blackHole":[`,
		} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("hint output without a playable move still keeps the board", func(t *testing.T) {
		// Hole top 5, fan0 top 10 (not ±1) -> no recommendation.
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":10,"w":true}]],"ph":0}`
		out := p.HintOutput(bhState(t, js))
		assert.Contains(t, out, "blackhole.noHint")
		assert.NotContains(t, out, `"hint":`)
		assert.Contains(t, out, `"fans":[[`)
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
