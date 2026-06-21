//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// ssState builds a SimpleSimon in an arbitrary state via JSON restore.
func ssState(t *testing.T, js string) *domain.SimpleSimon {
	t.Helper()
	g := domain.NewDefaultSimpleSimon()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestSimpleSimonCuiPresenter_Output(t *testing.T) {
	p := new(presenter.SimpleSimonCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Simple Simon")
		assert.Contains(t, out, "完成スート")
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(ssState(t, `{"ph":1,"mc":42}`), nil), "ゲームクリア")
	})
	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(ssState(t, `{"ph":2}`), nil), "ゲームオーバー")
	})
	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(ssState(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		// A board where col1's 8♠ can move onto col0's 9♠.
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(ssState(t, js)), "HINT")
		// No hint once ended.
		assert.NotEmpty(t, p.HintOutput(ssState(t, `{"ph":2}`)))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultSimpleSimon()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
