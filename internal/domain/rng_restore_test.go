//go:build test

package domain

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEveryRestoredGameReseedsItsRng は「KV から復元するゲームは UnmarshalJSON で
// 乱数源を張り直す」という規約をソースレベルで強制する。
//
// **なぜソースを読むテストなのか。**この穴はゲームごとに 1 行の抜けで生まれ、
// 症状は Cloudflare Worker 上でしか出ない —— サーバーはインタラクタをメモリに
// 保持するので、ローカルのテストもCUIも全部通る。実際 Tonk は本番でクラッシュ
// する状態のまま出荷されていた (Easy 難易度で `g.rng.Intn` が nil を触る)。
// ゲームごとに個別のテストを書く運用では、次に足すゲームで同じ抜けが起きる。
//
// 規約: `rng *rand.Rand` フィールドを持ち、かつ UnmarshalJSON を実装している
// ドメインは、UnmarshalJSON の中で `g.rng = ...` すること。
func TestEveryRestoredGameReseedsItsRng(t *testing.T) {
	dir := "."
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	rngField := regexp.MustCompile(`\brng\s+\*rand\.Rand`)
	unmarshal := regexp.MustCompile(`(?s)func \(g \*\w+\) UnmarshalJSON.*?\n\}\n`)
	reseed := regexp.MustCompile(`g\.rng\s*=`)

	checked := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		s := string(src)
		if !rngField.MatchString(s) {
			continue
		}
		body := unmarshal.FindString(s)
		if body == "" {
			continue // 復元しないドメイン (equity エンジン等) は対象外
		}
		checked++
		require.True(t, reseed.MatchString(body),
			"%s: UnmarshalJSON が g.rng を張り直していない。Worker は復元後に SetRand を"+
				" 呼ばないので、シャッフル以外で rng を使う経路が nil で落ちる", name)
	}

	// **対象 0 件で素通りさせない。**正規表現が実装とずれたら、このガードは
	// 黙って何も見ずに通る (#4662 のレビューで学んだ形)。
	require.Greater(t, checked, 8, "対象ゲームが少なすぎる — 検出の正規表現が実装とずれている")
	t.Logf("checked %d restorable games with a struct rng", checked)
}
