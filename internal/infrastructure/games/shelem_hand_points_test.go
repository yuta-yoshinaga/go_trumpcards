package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestShelemHandPointsMatchTheDomain guards the "out of 100" the Web prints.
//
// `ShelemPage.tsx` shows the defenders' points against the pool so a player can
// tell whether the contract is being stopped (#5754). The pool size is a domain
// constant; a literal in the page would keep reading 100 even if the card-point
// table changed, and every existing test would still pass.
func TestShelemHandPointsMatchTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "ShelemPage.tsx")
	src, err := os.ReadFile(path) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`const SHELEM_HAND_POINTS = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("ShelemPage.tsx no longer states SHELEM_HAND_POINTS as a literal")
	}
	got, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatalf("parse SHELEM_HAND_POINTS: %v", err)
	}
	if got != domain.ShelemHandPoints {
		t.Errorf("the page says %d card points are on the table, the domain says %d",
			got, domain.ShelemHandPoints)
	}
}
