package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestBoliviaWebDefaultPointLimitMatchesTheDomain pins the Web GUI's default
// target score to the domain's.
//
// The Web hook sends its config on every reset, so a stale default there is not
// cosmetic: it makes the browser play a different game from the CLI. The value
// was cloned from Samba as 10000 while Bolivia's domain default is 15000, and
// nothing compared the two — the docs, the manuals and the CUI all said 15000
// while the Web GUI quietly played to 10000.
func TestBoliviaWebDefaultPointLimitMatchesTheDomain(t *testing.T) {
	path := filepath.Join(repoRoot, "frontend/src/hooks/useBoliviaGame.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`(?s)DEFAULT_BOLIVIA_CONFIG: BoliviaConfig = \{.*?pointLimit:\s*(\d+)`).
		FindSubmatch(data)
	if m == nil {
		t.Fatal("useBoliviaGame.ts no longer states DEFAULT_BOLIVIA_CONFIG.pointLimit as a literal — " +
			"update this guard rather than dropping it")
	}
	got, err := strconv.Atoi(string(m[1]))
	if err != nil {
		t.Fatalf("parse pointLimit %q: %v", m[1], err)
	}
	if got != domain.BoliviaDefaultPointLimit {
		t.Errorf("useBoliviaGame.ts defaults pointLimit to %d, domain default is %d — "+
			"the Web GUI would play to a different target than the CLI",
			got, domain.BoliviaDefaultPointLimit)
	}
}
