//go:build test && (!js || !wasm || casino)

package presenter_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// dramahaRawKey は画面に出てはいけない生のキー (`dramaha.scoop` など)。
// 訳文には `.` を挟んだ英小文字の連なりは出てこない。
var dramahaRawKey = regexp.MustCompile(`dramaha\.[a-zA-Z][a-zA-Z0-9_]*`)

// TestDramahaCuiPresenter_NoRawKeyReachesThePlayer は、ロケールの取りこぼしが
// 画面に出ないことを見る。
//
// **`i18n.T(key)` を期待値にしたアサーションはこれを絶対に捕まえない。**
// `T` は未翻訳のときキー自身を返すので、キーを消すと presenter の出力も
// 期待値も揃って `dramaha.scoop` になり、`Contains` は通ったままになる。
// 実測: `dramaha.scoop` をロケールから消すと CUI に生キーが出るのに、
// `TestDramahaCuiPresenter_Scoop` は緑のままだった。
//
// なので期待値をキーから作るのをやめ、**出力に生キーが 1 つも無いこと**を
// 直接見る。ロケールを消せばこのテストだけが落ちる。
func TestDramahaCuiPresenter_NoRawKeyReachesThePlayer(t *testing.T) {
	p := new(presenter.DramahaCuiPresenter)

	// スクープ・Hi 片取り・Lo 片取り・負け —— 結果行の分岐を全部通す。
	resultSets := map[string][]domain.HoldemResult{
		"scoop": {{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
			WonAmount: 100, HiWonAmount: 60, LowWonAmount: 40}},
		"hi only": {{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
			WonAmount: 60, HiWonAmount: 60}},
		"lo only": {{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
			WonAmount: 40, LowWonAmount: 40}},
		"lost": {{PlayerIdx: 0, HandRank: domain.PokerHandHighCard, HandName: "High Card"}},
	}

	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)

			seen := 0
			check := func(what, out string) {
				seen++
				assert.Empty(t, dramahaRawKey.FindAllString(out, -1),
					"%s: 生のロケールキーが画面に出ている:\n%s", what, out)
				assert.NotContains(t, out, "{{",
					"%s: プレースホルダが素通りしている:\n%s", what, out)
			}

			for _, phase := range []int{
				domain.DramahaPhasePreFlop, domain.DramahaPhaseFlop,
				domain.DramahaPhaseDraw, domain.DramahaPhaseTurn,
				domain.DramahaPhaseRiver, domain.DramahaPhaseShowdown,
				domain.DramahaPhaseEnd,
			} {
				h := makeHeadsUpDramahaForPresenter()
				h.SetPhase(phase)
				check("phase render", p.Output(h, nil))
			}

			for name, results := range resultSets {
				h := makeHeadsUpDramahaForPresenter()
				h.SetPhase(domain.DramahaPhaseEnd)
				h.SetRoundResults(results)
				check(name, p.Output(h, nil))
			}

			require.Positive(t, seen, "1 画面も描画していない")
		})
	}
}

// TestDramahaRawKeyGuardSeesARawKey は上のガードの負のコントロール。
// **正規表現が何にも当たらないなら、上のテストは常に緑で何も見ていない。**
func TestDramahaRawKeyGuardSeesARawKey(t *testing.T) {
	assert.NotEmpty(t, dramahaRawKey.FindAllString("won: dramaha.scoop", -1),
		"ガードが生キーを見つけられない")
	assert.Empty(t, dramahaRawKey.FindAllString("スクープ! 100 (Hi 60 / Lo 40)", -1),
		"ガードが通常の訳文を誤検出している")
	assert.Empty(t, dramahaRawKey.FindAllString(strings.Repeat("ドラマハ ", 3), -1))
}
