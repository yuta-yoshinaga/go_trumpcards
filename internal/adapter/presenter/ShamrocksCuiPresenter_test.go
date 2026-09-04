//go:build test

package presenter_test

import (
	"encoding/json"
	"fmt"
	"strings"
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

// **動かせる扇に印が付く (#6592)。** Web は #5678 以降リングで常時示しているのに、
// CUI は hint を打つか、打って拒否されるまで分からなかった。
//
// 盤は JSON で組む。**空の扇を作らない** — Shamrocks は空の扇がどの札でも
// 受けるので、1 つでもあると全部が「動かせる」になって否定側を測れない。
func TestShamrocksCuiPresenter_MarksMovableFans(t *testing.T) {
	// 扇 0: K / 扇 1: 5 / 扇 2: A。**値を隣り合わせない** —
	// 積める条件は「差が 1」でスートを見ないので、2 を置くと A の隣になってしまう。
	// K と 5 はどこにも積めず、A だけが空の組札へ行ける。
	board := `{"ph":0,"mc":0,"fn":[` +
		`[{"d":0,"v":13,"f":true}],` +
		`[{"d":0,"v":5,"f":true}],` +
		`[{"d":0,"v":1,"f":true}]]}`
	out := new(presenter.ShamrocksCuiPresenter).Output(llStateSH(t, board), nil)

	lines := map[int]string{}
	for _, line := range strings.Split(out, "\n") {
		for i := 0; i < 3; i++ {
			if strings.HasPrefix(line, i18n.Tf("shamrocks.fanLabel", "idx", fmt.Sprint(i))) {
				lines[i] = line
			}
		}
	}

	// **印の有無だけを見る。** 期待値を i18n から組み立てると、未訳でも通ってしまう。
	assert.NotContains(t, lines[0], presenter.CuiLegalMark, "置き先の無い扇に印は付かない")
	assert.NotContains(t, lines[1], presenter.CuiLegalMark, "置き先の無い扇に印は付かない")
	assert.Contains(t, lines[2], presenter.CuiLegalMark, "組札へ行ける扇に印が付く")
}
