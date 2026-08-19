package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestAluetteWebTricksToWinMatchesTheDomain guards the web's copy of the meine
// threshold against the value the domain actually settles on.
//
// `frontend/src/pages/AluettePage.tsx` spells the "three tricks take the meine"
// rule as a literal so the round-result box can name the winning team (#5714);
// the CUI prints the same conclusion from `AluetteTricksToWin`. Change the trick
// count in the domain and the page would keep crowning the old winner.
func TestAluetteWebTricksToWinMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "pages", "AluettePage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`ALUETTE_TRICKS_TO_WIN = (\d+)`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("AluettePage.tsx no longer states ALUETTE_TRICKS_TO_WIN as a literal")
	}
	if m[1] != strconv.Itoa(domain.AluetteTricksToWin) {
		t.Errorf("web needs %s tricks, domain needs %d", m[1], domain.AluetteTricksToWin)
	}
}
