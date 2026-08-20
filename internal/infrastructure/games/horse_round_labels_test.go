//go:build test

package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestHorseRoundLabelsMatchTheDisciplinePhases は H.O.R.S.E. の画面が出す
// ラウンド名の並びが、実際に動いている種目のフェーズ番号と一致することを見る。
//
// **同じ 2 でもホールデムはフロップ、スタッドは 4th street。** ページ側の配列が
// ずれると、生の数字を出すより悪い「もっともらしい嘘」になる (#5788)。
func TestHorseRoundLabelsMatchTheDisciplinePhases(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "HorsePage.tsx")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("HorsePage.tsx を読めない: %v", err)
	}
	src := string(data)

	community := extractHorseRoundKeys(t, src, "COMMUNITY_ROUND_KEYS")
	stud := extractHorseRoundKeys(t, src, "STUD_ROUND_KEYS")

	// コミュニティ系はホールデムのフェーズ番号がそのまま添字になる。
	wantCommunity := map[int]string{
		domain.HoldemPhasePreFlop:  "preflop",
		domain.HoldemPhaseFlop:     "flop",
		domain.HoldemPhaseTurn:     "turn",
		domain.HoldemPhaseRiver:    "river",
		domain.HoldemPhaseShowdown: "showdown",
	}
	assertHorseRoundKeys(t, "COMMUNITY_ROUND_KEYS", community, wantCommunity)

	// スタッド系 (Razz / Stud / Stud Hi-Lo) は SevenCardStud のフェーズ番号。
	wantStud := map[int]string{
		domain.SevenCardStudPhaseThirdStreet:   "third",
		domain.SevenCardStudPhaseFourthStreet:  "fourth",
		domain.SevenCardStudPhaseFifthStreet:   "fifth",
		domain.SevenCardStudPhaseSixthStreet:   "sixth",
		domain.SevenCardStudPhaseSeventhStreet: "seventh",
		domain.SevenCardStudPhaseShowdown:      "showdown",
	}
	assertHorseRoundKeys(t, "STUD_ROUND_KEYS", stud, wantStud)

	// 添字 0 は「まだ配っていない」。ラベルを付けると初期状態が誤って名前を持つ。
	for _, name := range []string{"COMMUNITY_ROUND_KEYS", "STUD_ROUND_KEYS"} {
		keys := community
		if name == "STUD_ROUND_KEYS" {
			keys = stud
		}
		if keys[0] != "" {
			t.Errorf("%s[0] = %q, want empty (%d = 未開始)", name, keys[0], domain.HoldemPhaseInit)
		}
	}
}

// extractHorseRoundKeys は TSX の配列リテラルから要素を取り出す。
func extractHorseRoundKeys(t *testing.T, src, name string) []string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^const ` + name + ` = \[([^\]]*)\] as const;`)
	m := re.FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("%s が見つからない (名前を変えたならこのガードも直す)", name)
	}
	parts := strings.Split(m[1], ",")
	keys := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		keys = append(keys, strings.Trim(p, "'\""))
	}
	return keys
}

// assertHorseRoundKeys は添字ごとのラベルを突き合わせる。
func assertHorseRoundKeys(t *testing.T, name string, keys []string, want map[int]string) {
	t.Helper()
	for phase, key := range want {
		if phase >= len(keys) {
			t.Errorf("%s にフェーズ %d の要素が無い (%q のはず)", name, phase, key)
			continue
		}
		if keys[phase] != key {
			t.Errorf("%s[%d] = %q, want %q", name, phase, keys[phase], key)
		}
	}
}
