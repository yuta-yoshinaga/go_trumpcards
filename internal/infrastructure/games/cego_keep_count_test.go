package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestCegoWebKeepCountMatchesTheDomain guards the web's copy of the Cego keep
// count against the number the domain actually enforces.
//
// `frontend/src/hooks/useCegoGame.ts` spells it as a literal to gate the confirm
// button, and both the Web stepper and the CUI step lines are derived from it
// (#5718). Change `CegoKeepCount` and the page would keep demanding the old
// selection count.
func TestCegoWebKeepCountMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "hooks", "useCegoGame.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`CEGO_KEEP_COUNT = (\d+)`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("useCegoGame.ts no longer states CEGO_KEEP_COUNT as a literal")
	}
	if m[1] != strconv.Itoa(domain.CegoKeepCount) {
		t.Errorf("web keeps %s card(s), domain keeps %d", m[1], domain.CegoKeepCount)
	}
}
