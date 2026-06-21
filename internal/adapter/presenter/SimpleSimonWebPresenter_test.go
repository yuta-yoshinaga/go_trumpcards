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

	t.Run("hint output", func(t *testing.T) {
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(ssState(t, js)), "simplesimon.hintAvailable")
		assert.Contains(t, p.HintOutput(ssState(t, `{"ph":2}`)), "simplesimon.noHint")
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		assert.True(t, json.Valid([]byte(p.ActionLogOutput(g))))
	})
}
