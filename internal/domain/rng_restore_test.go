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

// TestNoDomainSeedsItsRngFromTheClock は「乱数の種に時計を使わない」規約を
// ソースレベルで強制する。
//
// **同じナノ秒に作った 2 局が同じ配りになる。**`time.Now().UnixNano()` は
// 時計の分解能が粗い環境や、Worker が短時間に複数セッションを立ち上げる状況で
// 衝突しうる。Go 1.20 以降グローバル `rand` は起動時にランダムに初期化される
// ので、`rand.Int63()` を種にすればこの衝突は起きない。
//
// **コメントは対象外。**「なぜ time を使わないか」を説明した行が数箇所あり、
// 素朴に文字列検索するとそれを実装だと誤認する (実際 Ganjifa / Minchiate /
// Tarocchini の 3 ファイルが該当する)。行頭のコメントを落としてから見る。
func TestNoDomainSeedsItsRngFromTheClock(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	lineComment := regexp.MustCompile(`(?m)^\s*//.*$`)
	clockSeed := regexp.MustCompile(`rand\.NewSource\(\s*time\.Now\(\)`)
	anySeed := regexp.MustCompile(`rand\.NewSource\(`)

	seeders := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		require.NoError(t, err)
		code := lineComment.ReplaceAllString(string(src), "")
		if !anySeed.MatchString(code) {
			continue
		}
		seeders++
		require.False(t, clockSeed.MatchString(code),
			"%s: 乱数の種が time.Now() —— 同一ナノ秒に構築された 2 局が同じ配りになる。"+
				"rand.NewSource(rand.Int63()) を使うこと (#4664)", name)
	}

	// **対象 0 件で素通りさせない。**正規表現が実装とずれたら、このガードは
	// 黙って何も見ずに通る。
	// 14 が現在の実測値。1 桁に落ちたら正規表現が実装から外れたと見てよい。
	require.Greater(t, seeders, 10, "種を張るドメインが少なすぎる — 検出の正規表現が実装とずれている")
	t.Logf("checked %d domains that seed an rng", seeders)
}
