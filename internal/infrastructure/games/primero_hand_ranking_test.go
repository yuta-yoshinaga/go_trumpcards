package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// primeroRankingFromDomain lists the Primero hand categories strongest first,
// as the domain constants order them.
func primeroRankingFromDomain() []string {
	byCategory := map[int]string{
		domain.PrimeroHandFluxus:   "fluxus",
		domain.PrimeroHandSupremus: "supremus",
		domain.PrimeroHandPrimero:  "primero",
		domain.PrimeroHandNumerus:  "numerus",
	}
	cats := []int{
		domain.PrimeroHandFluxus, domain.PrimeroHandSupremus,
		domain.PrimeroHandPrimero, domain.PrimeroHandNumerus,
	}
	// 強い順 = カテゴリ値の降順 (PrimeroCompare がカテゴリの大小で勝敗を決める)。
	for i := 0; i < len(cats); i++ {
		for j := i + 1; j < len(cats); j++ {
			if cats[j] > cats[i] {
				cats[i], cats[j] = cats[j], cats[i]
			}
		}
	}
	out := make([]string, 0, len(cats))
	for _, c := range cats {
		out = append(out, byCategory[c])
	}
	return out
}

// TestPrimeroWebHandRankingMatchesTheDomain guards the web's hand-legend order
// against the order the domain actually scores.
//
// `frontend/src/pages/PrimeroPage.tsx` spells the ranking as a literal array so
// its legend table can render rows strongest-first, and the CUI prints the same
// order from the domain constants (#5699). Swap two PrimeroHand* values and the
// legend would keep teaching the old order with nothing to notice.
func TestPrimeroWebHandRankingMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "pages", "PrimeroPage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`PRIMERO_HAND_RANKING = \[([^\]]*)\]`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("PrimeroPage.tsx no longer states PRIMERO_HAND_RANKING as a literal array")
	}
	got := make([]string, 0, 4)
	for _, part := range strings.Split(m[1], ",") {
		if name := strings.Trim(strings.TrimSpace(part), "'\""); name != "" {
			got = append(got, name)
		}
	}

	want := primeroRankingFromDomain()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("PRIMERO_HAND_RANKING = %v, domain order = %v", got, want)
	}
}
