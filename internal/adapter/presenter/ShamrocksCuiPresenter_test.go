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

// llStateSH builds a Shamrocks in an arbitrary state via JSON restore.
func llStateSH(t *testing.T, js string) *domain.Shamrocks {
	t.Helper()
	g := domain.NewDefaultShamrocks()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestShamrocksCuiPresenter_StuckSaysDeadlock(t *testing.T) {
	// La Belle Lucie's three redeal cases are gone: Shamrocks has none, so a
	// board with no legal move is always a true deadlock and there is no
	// "redeal recommended" state to reach. Assert against literal wording, not
	// against i18n.T of a key that no longer exists -- a deleted key makes T
	// return the key itself and NotContains would pass for the wrong reason.
	out := new(presenter.ShamrocksCuiPresenter).Output(llStateSH(t, `{"ph":0,"mc":0}`), nil)
	for _, frag := range []string{"再配り", "配り直", "シャッフル", "リディール"} {
		assert.NotContains(t, out, frag, "no redeal wording in a game without redeals")
	}
}

func TestShamrocksCuiPresenter_Output(t *testing.T) {
	p := new(presenter.ShamrocksCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Shamrocks")
		// La Belle Lucie printed a "残りシャッフル" line; Shamrocks has no redeals,
		// so that line is gone.
		assert.NotContains(t, out, "残りシャッフル")
		// What matters here is the 18-fan layout: 17 fans of three plus one of one.
		assert.Contains(t, out, "扇17:")
		assert.NotContains(t, out, "扇18:")
	})

	// **再配札が尽きた真の手詰まりは別物 (#4769)。**Web は ll-deadlock-banner を
	// 出して giveup を点滅させるのに、CUI は何も言わず、合法手が無いまま延々と
	// 手を探させていた。
	t.Run("names the deadlock when no legal move remains", func(t *testing.T) {
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"fd":[[%s],[],[],[]],"fn":[[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		out := p.Output(llStateSH(t, js), nil)
		assert.Contains(t, out, i18n.T("shamrocks.stuckDeadlock"))
	})

	t.Run("stays quiet while a legal move exists", func(t *testing.T) {
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"fd":[[],[],[],[]],"fn":[[%s],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		out := p.Output(llStateSH(t, js), nil)
		assert.NotContains(t, out, i18n.T("shamrocks.stuckDeadlock"))
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(llStateSH(t, `{"ph":1,"mc":7}`), nil), "ゲームクリア")
	})

	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(llStateSH(t, `{"ph":2}`), nil), "ゲームオーバー")
	})

	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(llStateSH(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		assert.NotEmpty(t, p.HintOutput(g))
		// Foundation hint for an exposed Ace.
		ace := llStateSH(t, `{"fn":[[{"d":1,"v":1,"w":true}]],"ph":0}`)
		assert.Contains(t, p.HintOutput(ace), "HINT")
		// No hint once the game has ended.
		assert.Contains(t, p.HintOutput(llStateSH(t, `{"ph":2}`)), "")
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultShamrocks()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
