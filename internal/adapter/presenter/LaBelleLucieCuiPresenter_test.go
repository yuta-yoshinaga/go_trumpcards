//go:build test

package presenter_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// llState builds a LaBelleLucie in an arbitrary state via JSON restore.
func llState(t *testing.T, js string) *domain.LaBelleLucie {
	t.Helper()
	g := domain.NewDefaultLaBelleLucie()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestLaBelleLucieCuiPresenter_Output(t *testing.T) {
	p := new(presenter.LaBelleLucieCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "La Belle Lucie")
		assert.Contains(t, out, "残りシャッフル")
	})

	t.Run("compact foundation and redeal recommendation when stuck", func(t *testing.T) {
		// Foundation 0 holds an Ace; all fans are empty, so there is no legal
		// move while redeals remain.
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"rd":2,"fd":[[%s],[],[],[]],"fn":[[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		g := llState(t, js)
		out := p.Output(g, nil)
		// Compact top-card + count for the non-empty foundation.
		assert.Contains(t, out, i18n.Tf("labellelucie.foundationCount", "count", "1"))
		// Empty foundations still show the shared empty marker.
		assert.Contains(t, out, i18n.T("cuiEmptyCol"))
		// No legal move + redeals remaining -> recommendation shown.
		assert.Contains(t, out, i18n.T("labellelucie.redealRecommended"))
	})

	t.Run("no redeal recommendation when a legal move exists", func(t *testing.T) {
		// An Ace on a fan top can always start a foundation -> legal move exists.
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"rd":2,"fd":[[],[],[],[]],"fn":[[%s],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		g := llState(t, js)
		out := p.Output(g, nil)
		assert.NotContains(t, out, i18n.T("labellelucie.redealRecommended"))
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(llState(t, `{"ph":1,"mc":7}`), nil), "ゲームクリア")
	})

	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(llState(t, `{"ph":2}`), nil), "ゲームオーバー")
	})

	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(llState(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		assert.NotEmpty(t, p.HintOutput(g))
		// Foundation hint for an exposed Ace.
		ace := llState(t, `{"fn":[[{"d":1,"v":1,"w":true}]],"ph":0}`)
		assert.Contains(t, p.HintOutput(ace), "HINT")
		// No hint once the game has ended.
		assert.Contains(t, p.HintOutput(llState(t, `{"ph":2}`)), "")
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultLaBelleLucie()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
