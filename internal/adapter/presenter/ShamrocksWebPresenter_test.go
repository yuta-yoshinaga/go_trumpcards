//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestShamrocksWebPresenter_Output(t *testing.T) {
	p := new(presenter.ShamrocksWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		out := p.Output(g, nil)
		// Shamrocks has no redeal, so the field La Belle Lucie ships must not
		// appear on the wire either -- a client reading it would show "0 left"
		// for a counter that never existed.
		assert.NotContains(t, out, `"redealsLeft"`)
		assert.Contains(t, out, `"legalTargets"`)
		for _, frag := range []string{`"fans"`, `"foundation"`, `"canUndo"`, `"messageCode":"shamrocks.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("hint output includes targets and ended empty targets", func(t *testing.T) {
		g := llStateSH(t, `{"fn":[[{"d":1,"v":1,"w":true}]],"ph":0}`)
		assert.Contains(t, p.HintOutput(g), `"legalTargets":{"fan-0":{"fan":[],"foundation":[0]}}`)
		assert.Contains(t, p.HintOutput(llStateSH(t, `{"ph":2}`)), `"legalTargets":{}`)
	})

	t.Run("nil destination lists serialize as empty arrays in both outputs", func(t *testing.T) {
		g := new(interfaces.MockShamrocksGame)
		g.On("GetPhase").Return(domain.ShamrocksPhasePlaying)
		g.On("GetMoveCount").Return(0)
		g.On("CanUndo").Return(false)
		g.On("GetFans").Return([][]*domain.Card{})
		g.On("GetFoundation").Return([domain.ShamrocksFoundationCnt][]*domain.Card{})
		g.On("GetMovableFans").Return([]int{0, 1})
		g.On("LegalTargets", 0).Return([]int(nil), []int{2})
		g.On("LegalTargets", 1).Return([]int{3}, []int(nil))
		g.On("GetHint").Return((*domain.ShamrocksHint)(nil))
		for _, output := range []string{p.Output(g, nil), p.HintOutput(g)} {
			assert.Contains(t, output, `"fan-0":{"fan":[],"foundation":[2]}`)
			assert.Contains(t, output, `"fan-1":{"fan":[3],"foundation":[]}`)
		}
		g.AssertExpectations(t)
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("game clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(llStateSH(t, `{"ph":1,"mc":9}`), nil), "shamrocks.gameClear")
		assert.Contains(t, p.Output(llStateSH(t, `{"ph":2}`), nil), "shamrocks.gameOver")
	})

	// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	t.Run("output carries the hint", func(t *testing.T) {
		// 扇の一番上がスペードのエース。台札を開始できるのでヒントが必ず出る。
		js := `{"fn":[[{"d":1,"v":1,"w":true}]],"ph":0}`
		assert.Contains(t, p.Output(llStateSH(t, js), nil), `"hint"`,
			"Output must carry the hint -- the frontend reads state.hint")
		assert.NotContains(t, p.Output(llStateSH(t, `{"ph":2}`), nil), `"hint"`)
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		assert.Contains(t, p.HintOutput(g), "shamrocks.")
		// No hint once ended.
		assert.Contains(t, p.HintOutput(llStateSH(t, `{"ph":2}`)), "shamrocks.noHint")
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
