package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// klaverjasFaceOf は表示用の呼び名。A/J/Q/K は数字ではなく文字で書かれている。
func klaverjasFaceOf(value int) string {
	switch value {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	default:
		return strconv.Itoa(value)
	}
}

// TestKlaverjasRankHelpMatchesTheDomain guards the two strength tables that
// exist in three places: the domain's own ordering and scoring, the CUI prompt
// text (`internal/i18n/locales/*/klaverjas.json`, `rankHelp*`) and the Web
// legend (`KLAVERJAS_TRUMP_ROWS` / `KLAVERJAS_PLAIN_ROWS` in
// `frontend/src/pages/KlaverjasPage.tsx`).
//
// The CUI text landed in #5831 and the Web tables in #4757, both spelled out by
// hand. Nothing checked either against the code that actually decides tricks, so
// a change to `trumpStrength` or `cardPoints` would leave two screens confidently
// teaching the wrong order (#5645).
func TestKlaverjasRankHelpMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	for _, tc := range []struct {
		name    string
		isTrump bool
		locKey  string
		tsVar   string
	}{
		{"trump", true, "rankHelpTrump", "KLAVERJAS_TRUMP_ROWS"},
		{"plain", false, "rankHelpPlain", "KLAVERJAS_PLAIN_ROWS"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := domain.KlaverjasRankTable(tc.isTrump)
			if len(rows) != 8 {
				t.Fatalf("rank table has %d rows, want 8", len(rows))
			}

			// --- CUI: 強い順の並びと、0 点でないカードの配点が本文にあること ---
			for _, loc := range []string{"ja", "en"} {
				path := filepath.Join(root, "internal", "i18n", "locales", loc, "klaverjas.json")
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				var d map[string]any
				if err := json.Unmarshal(raw, &d); err != nil {
					t.Fatalf("parse %s: %v", path, err)
				}
				line, ok := d[tc.locKey].(string)
				if !ok {
					t.Fatalf("%s %s missing", loc, tc.locKey)
				}
				// **並びの部分だけを見る。**行全体を順に走査すると、後半の配点欄
				// (「9=14」など) の数字が並びの証拠として拾われ、順序を入れ替えても
				// 通ってしまう (実際にこの書き方で誤って通した)。
				order := klaverjasOrderSegment(t, line)
				want := make([]string, len(rows))
				for i, r := range rows {
					want[i] = klaverjasFaceOf(r.Value)
				}
				if strings.Join(order, " > ") != strings.Join(want, " > ") {
					t.Errorf("%s %s order = %v, want %v", loc, tc.locKey, order, want)
				}
				for _, r := range rows {
					if r.Points == 0 {
						continue
					}
					want := klaverjasFaceOf(r.Value) + "=" + strconv.Itoa(r.Points)
					if !strings.Contains(line, want) {
						t.Errorf("%s %s = %q, want it to state %q", loc, tc.locKey, line, want)
					}
				}
			}

			// --- Web: 同じ並びと配点が KlaverjasPage.tsx の表にあること ---
			page := filepath.Join(root, "frontend", "src", "pages", "KlaverjasPage.tsx")
			src, err := os.ReadFile(page)
			if err != nil {
				t.Fatalf("read %s: %v", page, err)
			}
			block := klaverjasRowsBlock(t, string(src), tc.tsVar)
			got := regexp.MustCompile(`face: '([^']+)', points: (\d+)`).FindAllStringSubmatch(block, -1)
			if len(got) != len(rows) {
				t.Fatalf("%s has %d rows, want %d", tc.tsVar, len(got), len(rows))
			}
			for i, r := range rows {
				if got[i][1] != klaverjasFaceOf(r.Value) || got[i][2] != strconv.Itoa(r.Points) {
					t.Errorf("%s[%d] = %s/%s, want %s/%d",
						tc.tsVar, i, got[i][1], got[i][2], klaverjasFaceOf(r.Value), r.Points)
				}
			}
		})
	}
}

// klaverjasOrderSegment は「切り札の強さ: J > 9 > ...」の並びだけを取り出す。
// 配点欄 (全角/半角どちらの括弧でも) より前を、">" で切る。
func klaverjasOrderSegment(t *testing.T, line string) []string {
	t.Helper()
	seg := line
	if i := strings.IndexAny(seg, "（("); i >= 0 {
		seg = seg[:i]
	}
	if i := strings.Index(seg, ":"); i >= 0 {
		seg = seg[i+1:]
	}
	parts := strings.Split(seg, ">")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// klaverjasRowsBlock は const <name> ... ]; の中身を切り出す。
func klaverjasRowsBlock(t *testing.T, src, name string) string {
	t.Helper()
	start := strings.Index(src, "const "+name)
	if start < 0 {
		t.Fatalf("%s not found in KlaverjasPage.tsx", name)
	}
	end := strings.Index(src[start:], "];")
	if end < 0 {
		t.Fatalf("%s is not terminated", name)
	}
	return src[start : start+end]
}
