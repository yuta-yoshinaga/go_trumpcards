//go:build test

package games_test

import (
	"regexp"
	"strconv"
	"testing"
)

// statesNumber reports whether s states n as a number in its own right.
//
// **部分一致では通っても何も保証しない。** `strings.Contains(line, "0")` は
// `"10点"` の 0 でも通るので、期待値を 1 桁変えても落ちないガードになる
// (#6009)。数字の前後が別の数字でないことまで見る。
func statesNumber(s string, n int) bool {
	return numberPattern(strconv.Itoa(n)).MatchString(s)
}

// statesText reports whether s states the given numeric text (e.g. "0.95") in
// its own right. Same rule as statesNumber, for values that are not plain ints.
func statesText(s, num string) bool {
	return numberPattern(num).MatchString(s)
}

// numberPattern は「前後が数字でない」ことを要求する正規表現を返す。
//
// 小数点を含む値 (0.95) もそのまま扱えるよう、区切りの判定は数字だけで行い、
// 小数点は数値の一部として扱う。
func numberPattern(num string) *regexp.Regexp {
	return regexp.MustCompile(`(?:^|[^0-9.])` + regexp.QuoteMeta(num) + `(?:[^0-9.]|$)`)
}

// **ガード自身のガード。** 桁の一部で通ってしまう形に戻したら、ここが落ちる。
func TestStatesNumberRequiresADigitBoundary(t *testing.T) {
	cases := []struct {
		name string
		text string
		n    int
		want bool
	}{
		{"exact", "0点", 0, true},
		{"inside a longer number", "10点", 0, false},
		{"leading digit", "120点", 12, false},
		{"trailing digit", "612点", 12, false},
		{"among other text", "合計 120 点のうち 61 点で勝ち", 61, true},
		{"decimal is not a boundary", "0.95:1", 0, false},
		{"the whole decimal matches", "0.95:1", 95, false},
		{"at the start", "61 点", 61, true},
		{"at the end", "勝利ラインは 61", 61, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statesNumber(tc.text, tc.n); got != tc.want {
				t.Errorf("statesNumber(%q, %d) = %v, want %v", tc.text, tc.n, got, tc.want)
			}
		})
	}

	// 小数はそのまま照合できる。
	if !statesText("バンカー勝ち (0.95:1)", "0.95") {
		t.Error(`statesText("… (0.95:1)", "0.95") = false, want true`)
	}
	if statesText("バンカー勝ち (10.95:1)", "0.95") {
		t.Error(`statesText("… (10.95:1)", "0.95") = true, want false`)
	}
}
