package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestBoliviaWebMinMeldMatchesTheDomain guards the web's copy of the initial-meld
// thresholds against the values the domain actually enforces.
//
// `frontend/src/utils/boliviaScore.ts` spells the 15/50/90/120 ladder as literals
// so the page can show "必要点数" while the human picks cards; the CUI prints the
// same number from `BoliviaMinimumMeldValue` (#5702). Move a threshold in the
// domain and the page would keep promising the old one, with the mismatch only
// surfacing as a rejected meld.
func TestBoliviaWebMinMeldMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "utils", "boliviaScore.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	body := regexp.MustCompile(`(?s)export function boliviaMinMeld\(cumulativeScore: number\): number \{(.*?)\n\}`).
		FindStringSubmatch(string(src))
	if body == nil {
		t.Fatal("boliviaScore.ts no longer defines boliviaMinMeld as a plain function")
	}
	// `if (cumulativeScore < N) return M;` plus a trailing `return M;`.
	branches := regexp.MustCompile(`if \(cumulativeScore < (-?\d+)\) return (\d+);`).FindAllStringSubmatch(body[1], -1)
	fallback := regexp.MustCompile(`\n\s*return (\d+);`).FindAllStringSubmatch(body[1], -1)
	if len(branches) == 0 || len(fallback) == 0 {
		t.Fatalf("boliviaMinMeld no longer states its thresholds as literals: %q", body[1])
	}

	for _, b := range branches {
		bound, err := strconv.Atoi(b[1])
		if err != nil {
			t.Fatalf("threshold %q: %v", b[1], err)
		}
		// 境界の 1 点下はこの分岐に入る値。
		if got := domain.BoliviaMinimumMeldValue(bound - 1); strconv.Itoa(got) != b[2] {
			t.Errorf("web says %s points below %d, domain says %d", b[2], bound, got)
		}
		// **境界そのものも見る。**値だけを比べると、同じ値が続く帯の中へ境界を
		// 動かした web の変更 (3000 -> 2000 など) が素通りする。
		if got := domain.BoliviaMinimumMeldValue(bound); strconv.Itoa(got) == b[2] {
			t.Errorf("web breaks at %d but the domain still requires %s points there", bound, b[2])
		}
	}
	last := fallback[len(fallback)-1][1]
	if got := domain.BoliviaMinimumMeldValue(1_000_000); strconv.Itoa(got) != last {
		t.Errorf("web's fallback is %s points, domain says %d", last, got)
	}
}
