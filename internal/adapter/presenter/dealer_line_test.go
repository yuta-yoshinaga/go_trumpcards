//go:build test

package presenter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **ディーラー行が席番号を英語で名乗っていた。**日本語ロケールの中に
// `ディーラー: Player {{idx}}` と英単語が埋め込まれ、実名ではなく座席の
// 生の添字が出ていた ── 同じ画面の他の行はすべて `cuiPlayerName` を
// 通しているのに、この行だけ 12 ゲームで揃って外れていた (#6470)。
//
// 1 ゲームずつ出力を組んで確かめるより、**ロケール側の形を全部数える**ほうが
// 取りこぼしが無い: `{{idx}}` を受ける限り、渡せるのは番号だけになる。
func TestDealerLineNamesThePlayerNotTheSeat(t *testing.T) {
	for _, loc := range []string{"ja", "en"} {
		t.Run(loc, func(t *testing.T) {
			dir := filepath.Join("..", "..", "i18n", "locales", loc)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)

			checked, named := 0, 0
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
				require.NoError(t, err)
				var m map[string]any
				require.NoError(t, json.Unmarshal(raw, &m), e.Name())

				line, ok := m["dealerLine"].(string)
				if !ok {
					continue
				}
				checked++
				// **席番号は渡さない。**`{{idx}}` を受ける限り、渡せるのは
				// 番号だけになる。
				assert.NotContains(t, line, "{{idx}}",
					"%s/%s: dealerLine takes a seat index; pass a player name instead", loc, e.Name())
				if loc == "ja" {
					assert.NotContains(t, line, "Player",
						"%s/%s: the Japanese string still carries an English word", loc, e.Name())
				}
				if strings.Contains(line, "{{name}}") {
					named++
				}
			}
			// **0 件で成功と読まれない。**キー名を変えたり置き場所が動いたら、
			// この walk は何も見つけずに緑になる。
			assert.GreaterOrEqual(t, checked, 15,
				"expected the scan to reach the locale files; it found only %d dealerLine keys", checked)
			// **席を名乗る側だけを数える。**カジノ系の `dealerLine` はディーラーの
			// **手札**を出す別物 ("ディーラー: {{cards}} = {{score}}") なので、
			// 全部に `{{name}}` を要求すると誤検知になる。名前で呼ぶ 12 ゲームが
			// 減っていないことだけを見る。
			assert.GreaterOrEqual(t, named, 13,
				"only %d dealerLine strings name the dealer; one was reverted to a seat index", named)
		})
	}
}
