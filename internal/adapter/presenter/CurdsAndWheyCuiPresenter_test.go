//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ssStateCW builds a CurdsAndWhey in an arbitrary state via JSON restore.
func ssStateCW(t *testing.T, js string) *domain.CurdsAndWhey {
	t.Helper()
	g := domain.NewDefaultCurdsAndWhey()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

func TestCurdsAndWheyCuiPresenter_Output(t *testing.T) {
	p := new(presenter.CurdsAndWheyCuiPresenter)

	t.Run("playing board", func(t *testing.T) {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Curds and Whey")
		assert.Contains(t, out, "完成スート")
	})

	t.Run("game clear banner", func(t *testing.T) {
		assert.Contains(t, p.Output(ssStateCW(t, `{"ph":1,"mc":42}`), nil), "ゲームクリア")
	})
	t.Run("game over banner", func(t *testing.T) {
		assert.Contains(t, p.Output(ssStateCW(t, `{"ph":2}`), nil), "ゲームオーバー")
	})
	t.Run("error block", func(t *testing.T) {
		assert.NotEmpty(t, p.Output(ssStateCW(t, `{"ph":0}`), assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		// A board where col1's 8♠ can move onto col0's 9♠.
		js := `{"co":[[{"d":1,"v":9,"w":true}],[{"d":1,"v":8,"w":true}]],"ph":0}`
		assert.Contains(t, p.HintOutput(ssStateCW(t, js)), "HINT")
		// No hint once ended.
		assert.NotEmpty(t, p.HintOutput(ssStateCW(t, `{"ph":2}`)))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultCurdsAndWhey()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}

// #5679: まとめて動かせるのは「末尾から同スート降順に続く塊 (run)」だけ。Web は
// その起点にリングを付けているのに、CUI は札を平らに並べるだけで、どこからが
// 掴めるのかを毎回目視で確かめさせていた。
func TestCurdsAndWheyCuiPresenter_ShowsTheRunBoundary(t *testing.T) {
	p := new(presenter.CurdsAndWheyCuiPresenter)

	// 列0: ♠K ♥5 ♥4 ♥3 — 末尾3枚 (♥5-4-3) が run。
	js := `{"co":[[{"d":1,"v":13,"w":true},{"d":3,"v":5,"w":true},` +
		`{"d":3,"v":4,"w":true},{"d":3,"v":3,"w":true}]],"ph":0}`

	t.Run("marks where the movable run starts", func(t *testing.T) {
		out := p.Output(ssStateCW(t, js), nil)

		// 区切りは run の直前に入る。♠K の後、♥5 の前。
		assert.Contains(t, out, presenter.CuiRunMark+" "+color.Red("HEART 5"))
		assert.NotContains(t, out, presenter.CuiRunMark+" "+color.Red("HEART 4"))
	})

	// **run が末尾1枚だけの列でも境界を示す。**「掴めるのは1枚だけ」が分かる。
	t.Run("marks a single-card run", func(t *testing.T) {
		single := `{"co":[[{"d":1,"v":13,"w":true},{"d":3,"v":2,"w":true}]],"ph":0}`
		out := p.Output(ssStateCW(t, single), nil)

		assert.Contains(t, out, presenter.CuiRunMark+" "+color.Red("HEART 2"))
	})

	// 列全体が run なら先頭に印は要らない (全部掴める)。
	t.Run("no marker when the whole column is one run", func(t *testing.T) {
		whole := `{"co":[[{"d":1,"v":5,"w":true},{"d":1,"v":4,"w":true}]],"ph":0}`
		out := p.Output(ssStateCW(t, whole), nil)

		assert.NotContains(t, out, presenter.CuiRunMark)
	})

	// **run は同スート限定。**ランクだけ続いていても、スートが違えば塊にならない。
	// ♠6 ♥5 ♥4 → 掴めるのは ♥5-4 の 2 枚だけ。
	t.Run("a rank-adjacent card of another suit does not extend the run", func(t *testing.T) {
		mixed := `{"co":[[{"d":1,"v":6,"w":true},{"d":3,"v":5,"w":true},` +
			`{"d":3,"v":4,"w":true}]],"ph":0}`
		out := p.Output(ssStateCW(t, mixed), nil)

		assert.Contains(t, out, presenter.CuiRunMark+" "+color.Red("HEART 5"))
		assert.NotContains(t, out, presenter.CuiRunMark+" SPADE 6")
	})

	// 空列の表示は変えない (受け入れ条件3)。
	t.Run("an empty column is unchanged", func(t *testing.T) {
		out := p.Output(ssStateCW(t, `{"co":[[]],"ph":0}`), nil)

		assert.Contains(t, out, i18n.T("cuiEmptyCol"))
	})
}
