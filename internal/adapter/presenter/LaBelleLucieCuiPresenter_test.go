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

	// **再配札が尽きた真の手詰まりは別物 (#4769)。**Web は ll-deadlock-banner を
	// 出して giveup を点滅させるのに、CUI は何も言わず、合法手が無いまま延々と
	// 手を探させていた。
	t.Run("names the deadlock when no redeal is left either", func(t *testing.T) {
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"rd":0,"fd":[[%s],[],[],[]],"fn":[[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		out := p.Output(llState(t, js), nil)
		assert.Contains(t, out, i18n.T("labellelucie.stuckDeadlock"))
		// **再配札を勧めてはいけない。**もう配り直せない。
		assert.NotContains(t, out, i18n.T("labellelucie.redealRecommended"))
	})

	t.Run("recommends a redeal rather than a giveup while redeals remain", func(t *testing.T) {
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"rd":2,"fd":[[%s],[],[],[]],"fn":[[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		out := p.Output(llState(t, js), nil)
		assert.NotContains(t, out, i18n.T("labellelucie.stuckDeadlock"))
	})

	t.Run("says neither while a legal move exists", func(t *testing.T) {
		ace, _ := json.Marshal(domain.NewCard(domain.CardDesignSpade, 1, true))
		js := fmt.Sprintf(`{"ph":0,"rd":0,"fd":[[],[],[],[]],"fn":[[%s],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[],[]]}`, ace)
		out := p.Output(llState(t, js), nil)
		assert.NotContains(t, out, i18n.T("labellelucie.stuckDeadlock"))
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

// **どの扇が動かせるかは押すまで分からなかった。**Web は #5678 で事前に印を
// 出したが CUI には何も無く、CUI 利用者は #5678 以前の Web と同じ状態に
// 置かれていた (#6474)。
func TestLaBelleLucieCuiPresenter_MarksTheMovableFans(t *testing.T) {
	p := new(presenter.LaBelleLucieCuiPresenter)

	// 扇 0 = ♠9 (行き先なし)、扇 1 = ♠8 (♠9 に載る)、扇 2 = ♦A (台へ行ける)。
	// 残りは空。ファウンデーションはすべて空なのでエースだけが上がれる。
	mk := func(d, v int) string {
		b, err := json.Marshal(domain.NewCard(d, v, true))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return string(b)
	}
	fans := fmt.Sprintf(`[[%s],[%s],[%s]`, mk(domain.CardDesignSpade, 9), mk(domain.CardDesignSpade, 8), mk(domain.CardDesignDiamond, 1))
	for i := 3; i < 18; i++ {
		fans += ",[]"
	}
	fans += "]"
	g := llState(t, fmt.Sprintf(`{"ph":0,"rd":2,"fd":[[],[],[],[]],"fn":%s}`, fans))

	// **ドメインの答えと画面の印は同じものであること。**ここで前提を確かめて
	// おかないと、以下の行照合が「たまたま一致した」だけになる。
	movable := g.GetMovableFans()
	assert.Equal(t, []bool{false, true, true}, movable[:3])

	out := p.Output(g, nil)
	line := func(idx int) string { return i18n.Tf("labellelucie.fanLabel", "idx", fmt.Sprint(idx)) }
	assert.Contains(t, out, line(1)+presenter.CuiLegalMark)
	assert.Contains(t, out, line(2)+presenter.CuiLegalMark)
	// 動かせない扇には印を付けない ── 全部に付ける実装はここで落ちる。
	assert.NotContains(t, out, line(0)+presenter.CuiLegalMark)
	assert.NotContains(t, out, line(3)+presenter.CuiLegalMark)
}
