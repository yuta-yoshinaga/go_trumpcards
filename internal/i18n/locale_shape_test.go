//go:build test

package i18n_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TestLocaleFilesAreFlat は **ネストした object を 1 つでも含むロケールファイルを
// 落とす**。loadTranslations は `map[string]string` に Unmarshal し、失敗したら
// `continue` するので、`{"handEnd": {"tens": "..."}}` を 1 箇所書いただけで
// **そのゲームの訳が丸ごと消え、全キーが生のまま画面に出る**。
//
// 症状が出ないのが厄介で、`assert.Contains(out, i18n.T("x.y"))` は両辺とも
// 生キー "x.y" になって素通しする。だから presenter のテストではなくここで見る。
func TestLocaleFilesAreFlat(t *testing.T) {
	checked := 0
	for _, lang := range []string{"ja", "en"} {
		dir := filepath.Join("locales", lang)
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(path) //nolint:gosec // テスト内の固定パス
			require.NoError(t, err)

			var flat map[string]string
			assert.NoError(t, json.Unmarshal(data, &flat),
				path+": 値はすべて文字列でなければならない（ネストした object があると訳が丸ごと消える）")
			checked++
		}
	}
	// **走査ツールは 0 件でも成功する。** 実際に読んだことを下限で固定する。
	assert.Greater(t, checked, 400, "ロケールファイルを実際に走査している")
}

// TestLocaleFilesAreFlat_NegativeControl はガードが正しい形を弾かず、
// 壊れた形をちゃんと弾くことを確かめる。
func TestLocaleFilesAreFlat_NegativeControl(t *testing.T) {
	var flat map[string]string
	assert.NoError(t, json.Unmarshal([]byte(`{"a":"1","b":"2"}`), &flat), "平坦なら通る")
	assert.Error(t, json.Unmarshal([]byte(`{"a":{"b":"1"}}`), &flat), "ネストしていたら落ちる")
}

// TestEveryLocaleFileContributesKeys は **読み捨てられたファイルが 1 つも無い**
// ことを、実際に解決できるキーの側から確かめる。各ファイルの先頭キーを 1 つ
// 引いて、生キーが返ってこないことを見る。
func TestEveryLocaleFileContributesKeys(t *testing.T) {
	i18n.SetLang("ja")
	entries, err := os.ReadDir(filepath.Join("locales", "ja"))
	require.NoError(t, err)

	global := map[string]bool{"common.json": true, "cui_common.json": true}
	checked := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("locales", "ja", e.Name())) //nolint:gosec // テスト内の固定パス
		require.NoError(t, err)
		var flat map[string]string
		if err := json.Unmarshal(data, &flat); err != nil {
			continue // 形の検査は TestLocaleFilesAreFlat の仕事
		}
		ns := strings.TrimSuffix(e.Name(), ".json")
		for k := range flat {
			key := ns + "." + k
			if global[e.Name()] {
				key = k
			}
			assert.NotEqual(t, key, i18n.T(key), e.Name()+" が読み込まれていない")
			checked++
			break
		}
	}
	assert.Greater(t, checked, 200, "ロケールファイルを実際に走査している")
}
