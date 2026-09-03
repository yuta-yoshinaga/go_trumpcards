//go:build test

package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// A `ラベル:` at the start of a line inside the 画面の見方 sample block.
var manualRowLabelRe = regexp.MustCompile(`(?m)^([\p{Han}\p{Hiragana}\p{Katakana}][\p{Han}\p{Hiragana}\p{Katakana}]{1,9}):`)

var manualSampleBlockRe = regexp.MustCompile("(?s)## 画面の見方\\s*\n+```\n(.*?)```")

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// TestManualSampleRowsExist keeps docs/manual/cui/*.md honest about the board
// they claim to show. 57 of them had drifted by 2026-09: rows that no game
// prints (`捨て札トップ:`, `現在のトリック:`, `フェーズ: BID`), invented
// vocabulary (`コンテ` where the code says 結婚宣言), and 1-indexed column
// numbers where the CUI is 0-indexed -- typing the documented number moved the
// neighbouring column (#7061).
//
// **Two sources, because each alone is wrong in a different direction.**
//
//   - Rendered output alone over-flags: a row that only appears in a later
//     phase is missing from a freshly-dealt board even though the manual is
//     right. `宣言者:` in solowhist proved that.
//   - Locale text alone over-flags too: presenters share keys across games, so
//     `コミュニティ:` is absent from irishpoker's own file while the code
//     really does print it.
//
// A label is reported only when **both** miss it. That intersection was
// measured at 53 before this batch and 0 after.
func TestManualSampleRowsExist(t *testing.T) {
	root := repoRootForHelpTest(t)
	i18n.SetLang("ja")

	locales := readAllJaLocales(t, root)
	rendered := renderEveryGame()

	labels := 0
	for _, entry := range gameRegistry {
		md := filepath.Join(root, "docs/manual/cui", entry.Name+".md")
		raw, err := os.ReadFile(md)
		if err != nil {
			continue // TestPerGameManualsMatchRegistry owns "the manual exists"
		}
		block := manualSampleBlockRe.FindSubmatch(raw)
		if block == nil {
			continue
		}
		board := rendered[entry.Name]
		for _, m := range manualRowLabelRe.FindAllSubmatch(block[1], -1) {
			label := string(m[1])
			labels++
			if strings.Contains(board, label+":") || strings.Contains(locales, label) {
				continue
			}
			t.Errorf("%s.md: 見本の %q という行は、実際の描画にもロケールにも無い。"+
				"実物を描画して見本ごと直すこと (#7061)", entry.Name, label+":")
		}
	}

	// **緑であることは「見た」ことを意味しない。** 実測 684。走査が壊れて
	// 0 件になったら、この試験は何も主張せずに通ってしまう。
	//
	// 負のコントロール: 見本ブロックの正規表現を別物に変えると 0 件になり、
	// ここで落ちる。
	if labels < 500 {
		t.Fatalf("行ラベルを %d 件しか見ていない -- 走査が壊れている", labels)
	}
}

// readAllJaLocales concatenates every ja locale file. Presenters legitimately
// reuse another game's keys, so the search is deliberately repo-wide here; the
// rendered board is what makes the check specific.
func readAllJaLocales(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "internal/i18n/locales/ja")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var b strings.Builder
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var probe map[string]any
		if json.Unmarshal(raw, &probe) != nil {
			continue
		}
		b.Write(raw)
	}
	return b.String()
}

// renderEveryGame drives each registered game far enough to draw a board. A
// game whose controller panics on an unknown command is skipped rather than
// failing the suite -- the locale half still covers it.
func renderEveryGame() map[string]string {
	out := make(map[string]string, len(gameRegistry))
	cmds := []string{"", "r", "n", "p", "d", "s", "h", "c", "k", "log"}
	for _, entry := range gameRegistry {
		func() {
			defer func() { _ = recover() }()
			ctrl := entry.NewCui().Controller()
			var b strings.Builder
			for _, c := range cmds {
				func() {
					defer func() { _ = recover() }()
					b.WriteString(ctrl.Exec(c))
					b.WriteString("\n")
				}()
			}
			out[entry.Name] = ansiRe.ReplaceAllString(b.String(), "")
		}()
	}
	return out
}
