//go:build test

package presenter_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// bhState builds a BlackHole in an arbitrary state via JSON restore.
func bhState(t *testing.T, js string) *domain.BlackHole {
	t.Helper()
	g := domain.NewDefaultBlackHole()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

// **±1 を暗算させない (#4818)。**Web は「出せるランク」と「残り合法手」を常時
// 出しているのに、CUI は穴のトップと扇一覧しか出していなかった。
func TestBlackHoleCuiPresenter_AcceptableRanksAndLegalMoves(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BlackHoleCuiPresenter)

	// 穴のトップは ♠7。扇 0 のトップ ♥8 は積める、扇 1 のトップ ♣2 は積めない。
	board := `{"ph":0,"bh":[{"d":1,"v":7}],"fn":[[{"d":3,"v":8}],[{"d":2,"v":2}]]}`
	out := p.Output(bhState(t, board), nil)
	assert.Contains(t, out, "出せるランク: 6 / 8")
	assert.Contains(t, out, "残り合法手: 1")

	// 端 (A) は片側だけ。
	edge := `{"ph":0,"bh":[{"d":1,"v":1}],"fn":[[{"d":3,"v":2}]]}`
	assert.Contains(t, p.Output(bhState(t, edge), nil), "出せるランク: 2")

	// 積める扇が無ければ 0。
	stuck := `{"ph":0,"bh":[{"d":1,"v":7}],"fn":[[{"d":2,"v":2}]]}`
	assert.Contains(t, p.Output(bhState(t, stuck), nil), "残り合法手: 0")

	// 終了フェーズには出さない。
	assert.NotContains(t, p.Output(bhState(t, `{"ph":1,"mc":51}`), nil), "出せるランク")
}

func TestBlackHoleCuiPresenter_Output(t *testing.T) {
	p := new(presenter.BlackHoleCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Black Hole")
		assert.Contains(t, out, "ブラックホール")
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(bhState(t, `{"ph":1,"mc":51}`), nil), "ゲームクリア")
	})
	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(bhState(t, `{"ph":2}`), nil), "ゲームオーバー")
	})
	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(bhState(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		// Black hole top = 5, a fan top = 6 -> playable hint.
		js := `{"bh":[{"d":1,"v":5,"w":true}],"fn":[[{"d":2,"v":6,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(bhState(t, js)), "HINT")
		assert.NotEmpty(t, p.HintOutput(bhState(t, `{"ph":2}`)))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultBlackHole()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}

// #5681: 勝利条件は 52 枚すべてを穴へ吸い込むこと。17 個の扇を掘り進める長い
// ゲームなのに、**あと何枚で終わるか**が Web にも CUI にも出ていなかった。
func TestBlackHoleCuiPresenter_ShowsProgress(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BlackHoleCuiPresenter)

	t.Run("counts the cards swallowed against the deck", func(t *testing.T) {
		board := `{"ph":0,"bh":[{"d":1,"v":7},{"d":3,"v":8}],"fn":[[{"d":2,"v":2}]]}`

		out := p.Output(bhState(t, board), nil)

		assert.Contains(t, out, i18n.Tf("blackhole.progress",
			"count", "2", "total", strconv.Itoa(domain.BlackHoleTotalCards)))
	})

	// **クリア直前と直後で数字が正しく動く。**残り1枚を出し切ったら 52/52。
	t.Run("reads 52 of 52 once the board is cleared", func(t *testing.T) {
		cards := make([]string, domain.BlackHoleTotalCards)
		for i := range cards {
			cards[i] = `{"d":1,"v":1}`
		}
		board := `{"ph":1,"bh":[` + strings.Join(cards, ",") + `],"fn":[]}`

		out := p.Output(bhState(t, board), nil)

		assert.Contains(t, out, i18n.Tf("blackhole.progress",
			"count", strconv.Itoa(domain.BlackHoleTotalCards),
			"total", strconv.Itoa(domain.BlackHoleTotalCards)))
	})
}
