package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestFrenchTarotWebEcartReasonsMatchTheDomain guards the web's set of écart
// rejection reasons against the identifiers the domain actually produces.
//
// `frontend/src/utils/frenchtarotEcart.ts` spells the reasons as a TypeScript
// union so the page can key a per-card tooltip off them, and the CUI prints the
// same rule from `FrenchTarotUnburiableReason` (#5712). Rename or add a reason on
// one side and the other keeps showing the old vocabulary with nothing to notice.
func TestFrenchTarotWebEcartReasonsMatchTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "utils", "frenchtarotEcart.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`export type FrenchTarotUnburiableReason = ([^;]+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("frenchtarotEcart.ts no longer declares FrenchTarotUnburiableReason as a union")
	}
	got := make([]string, 0, 4)
	for _, part := range strings.Split(m[1], "|") {
		if name := strings.Trim(strings.TrimSpace(part), "'\""); name != "" {
			got = append(got, name)
		}
	}
	sort.Strings(got)

	want := []string{
		domain.FrenchTarotUnburiableBout,
		domain.FrenchTarotUnburiableExcuse,
		domain.FrenchTarotUnburiableKing,
		domain.FrenchTarotUnburiableTrump,
	}
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("web reasons = %v, domain reasons = %v", got, want)
	}
}
