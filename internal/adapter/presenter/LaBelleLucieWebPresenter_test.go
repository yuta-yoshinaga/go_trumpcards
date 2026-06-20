//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestLaBelleLucieWebPresenter_Output(t *testing.T) {
	p := new(presenter.LaBelleLucieWebPresenter)

	t.Run("playing state serialises board", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"fans"`, `"foundation"`, `"redealsLeft"`, `"canUndo"`, `"messageCode":"labellelucie.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("error surfaced", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("game clear / over message codes", func(t *testing.T) {
		assert.Contains(t, p.Output(llState(t, `{"ph":1,"mc":9}`), nil), "labellelucie.gameClear")
		assert.Contains(t, p.Output(llState(t, `{"ph":2}`), nil), "labellelucie.gameOver")
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		assert.Contains(t, p.HintOutput(g), "labellelucie.")
		// No hint once ended.
		assert.Contains(t, p.HintOutput(llState(t, `{"ph":2}`)), "labellelucie.noHint")
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
