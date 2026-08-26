//go:build test && (!js || !wasm || extra)

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TestSpeculationCuiPresenter_NoPlaceholderReachesThePlayer pins the
// interpolation parameter names against the locale placeholders.
//
// **名前がずれても `i18n.Tf` はエラーを出さない。** 置換が起きず、テンプレートが
// そのまま返るだけ —— `アンテ: {{amount}}` が画面に出る形で壊れる。しかも
// 期待値を `i18n.Tf(...)` から作ると両辺が同じテンプレートになり、どんな
// アサーションも通ってしまう。**`{{` が 1 つも出ないこと**を直接見るのが唯一
// 空振りしない検査。
func TestSpeculationCuiPresenter_NoPlaceholderReachesThePlayer(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)

			p := new(SpeculationCuiPresenter)
			for _, phase := range []domain.SpeculationPhase{
				domain.SpeculationPhaseFlip,
				domain.SpeculationPhaseAuction,
				domain.SpeculationPhaseResult,
				domain.SpeculationPhaseGameEnd,
			} {
				g := domain.NewDefaultSpeculation()
				g.SetPhase(phase)
				g.SetOffer(1, 0, 25) // 競りの行も必ず通す
				out := p.Output(g, nil)
				assert.NotContains(t, out, "{{",
					"phase %d: プレースホルダが素通りしている:\n%s", phase, out)
				assert.NotEmpty(t, strings.TrimSpace(out))

				hint := p.HintOutput(g)
				assert.NotContains(t, hint, "{{", "phase %d のヒント: %s", phase, hint)
			}

			// 人間が買い手の向きも通す (別のキーを使う)。
			g := domain.NewDefaultSpeculation()
			g.SetPhase(domain.SpeculationPhaseAuction)
			g.SetOffer(0, 1, 25)
			assert.NotContains(t, p.Output(g, nil), "{{")
			assert.NotContains(t, p.HintOutput(g), "{{")
		})
	}
}

// TestSpeculationCuiPresenter_ShowsRealNumbers is the negative control for the
// test above: `{{` の否定だけだと、値が空でも通ってしまう。
func TestSpeculationCuiPresenter_ShowsRealNumbers(t *testing.T) {
	defer i18n.SetLang(i18n.Lang())
	i18n.SetLang("ja")

	cfg := domain.NewDefaultSpeculationConfig()
	cfg.Players, cfg.Stake, cfg.InitialChips, cfg.Rounds = 3, 10, 500, 7
	g := domain.NewSpeculation(cfg)
	require.Equal(t, 30, g.GetPot())

	out := new(SpeculationCuiPresenter).Output(g, nil)
	for _, want := range []string{"30", "490", "7"} {
		assert.Contains(t, out, want, "実際の数字が出ていること:\n%s", out)
	}
}

// TestSpeculationCuiPresenter_ResultShowsTheRoundThatJustEnded pins the round
// number on the result screen.
//
// roundNo は決着で 1 進むので、そのまま +1 すると 1 回戦の結果画面に
// 「ラウンド 2 / 5」と出て、まだ始めていない回の結果を見ているように読める。
func TestSpeculationCuiPresenter_ResultShowsTheRoundThatJustEnded(t *testing.T) {
	defer i18n.SetLang(i18n.Lang())
	i18n.SetLang("ja")
	p := new(SpeculationCuiPresenter)

	g := domain.NewDefaultSpeculation()
	g.SetRoundNo(0)
	g.SetPhase(domain.SpeculationPhaseFlip)
	assert.Contains(t, p.Output(g, nil), "ラウンド: 1 /", "進行中は次に戦う回")

	g.SetRoundNo(1) // 1 回戦が決着した直後
	g.SetPhase(domain.SpeculationPhaseResult)
	out := p.Output(g, nil)
	assert.Contains(t, out, "ラウンド: 1 /", "結果画面は今終わった回")
	assert.NotContains(t, out, "ラウンド: 2 /")
}
