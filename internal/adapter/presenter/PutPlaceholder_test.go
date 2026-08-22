//go:build test && (!js || !wasm || extra4)

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TestPutCuiPresenter_NoPlaceholderReachesThePlayer pins the interpolation
// parameter names against the locale placeholders.
//
// **名前がずれても `i18n.Tf` はエラーを出さない。** 置換が起きず、テンプレートが
// そのまま返るだけ —— 実測で `トリック: {{trick}}` が画面に出た (ロケールの
// 用語を バサ→トリック に直したとき、プレゼンタが渡す `"baza"` を直し忘れた)。
//
// 期待値を `i18n.Tf(...)` から作ると両辺が同じテンプレートになって必ず通るので、
// **`{{` が 1 つも出ないこと**を直接見る。生のキー (`put.xxx`) も同様に、
// ロケールの取りこぼしがそのまま画面に出た印なので弾く。
func TestPutCuiPresenter_NoPlaceholderReachesThePlayer(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			defer i18n.SetLang(i18n.Lang())
			i18n.SetLang(lang)

			p := new(PutCuiPresenter)
			rendered := 0
			for _, phase := range []domain.PutPhase{
				domain.PutPhasePlay,
				domain.PutPhaseRespond,
				domain.PutPhaseTrickEnd,
				domain.PutPhaseHandEnd,
				domain.PutPhaseGameEnd,
			} {
				g := domain.NewDefaultPut()
				require.NotNil(t, g)
				g.Reset()
				g.SetPhase(phase)

				out := p.Output(g, nil)
				rendered++
				assert.NotContains(t, out, "{{",
					"phase %d: プレースホルダが素通りしている:\n%s", phase, out)
				assert.NotContains(t, out, "put.",
					"phase %d: 生のロケールキーが画面に出ている:\n%s", phase, out)
				assert.NotEmpty(t, strings.TrimSpace(out))

				hint := p.HintOutput(g)
				assert.NotContains(t, hint, "{{", "phase %d のヒント: %s", phase, hint)
				assert.NotContains(t, hint, "put.", "phase %d のヒント: %s", phase, hint)
			}
			require.Equal(t, 5, rendered, "5 フェーズすべてを描画していない")
		})
	}
}

// TestPutPlaceholderGuardSeesAPlaceholder は上のガードの負のコントロール。
// **`{{` を含む文字列で落ちないなら、上のテストは何も見ていない。**
func TestPutPlaceholderGuardSeesAPlaceholder(t *testing.T) {
	assert.True(t, strings.Contains("トリック: {{trick}}", "{{"),
		"検査対象の部分文字列が変わっている")
	assert.False(t, strings.Contains("トリック: 1", "{{"))
}
