package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestTrexPenaltyMirrorMatchesTheDomain keeps the Web's copy of the penalty
// rule honest.
//
// The CUI now calls domain.TrexCardPenalty directly, so it cannot drift. The
// Web cannot -- `frontend/src/utils/trexPenaltyCards.ts` re-implements the same
// switch in TypeScript, and its own doc comment says so. Nothing enforced that
// claim: change which cards a contract punishes and the red rings keep pointing
// at the old ones, on the surface where the rule is hardest to hold in mind.
//
// The check is deliberately coarse -- it reads the mirror's source and asserts
// each contract's condition is still spelled the way the domain scores it --
// because a Go test cannot execute TypeScript. It catches a rule being changed
// on one side, which is the failure that actually happens.
func TestTrexPenaltyMirrorMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "utils", "trexPenaltyCards.ts")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(raw)

	// The domain's own answers, asked of the cards each contract cares about.
	king := domain.NewCard(domain.CardDesignHeart, 13, true)
	queen := domain.NewCard(domain.CardDesignClover, 12, true)
	diamond := domain.NewCard(domain.CardDesignDiamond, 5, true)

	for _, tc := range []struct {
		name   string
		want   bool
		mirror string
	}{
		{"king of hearts", domain.TrexCardPenalty(domain.TrexContractKingOfHearts, king) != 0,
			`card.design === 'HEART' && card.value === 13`},
		{"diamonds", domain.TrexCardPenalty(domain.TrexContractDiamonds, diamond) != 0,
			`card.design === 'DIAMOND'`},
		{"queens", domain.TrexCardPenalty(domain.TrexContractQueens, queen) != 0,
			`card.value === 12`},
	} {
		if !tc.want {
			t.Fatalf("%s: the domain no longer penalises this card -- update the mirror and this test", tc.name)
		}
		if !strings.Contains(src, tc.mirror) {
			t.Errorf("%s: %s no longer spells %q; the Web rings and the score would disagree",
				tc.name, path, tc.mirror)
		}
	}

	// 個別札の減点が無い契約は false を返し続けること。default 節が消えると、
	// Tricks / Trix で無関係な札に印が付く。
	if !regexp.MustCompile(`default:\s*\n\s*return false;`).MatchString(src) {
		t.Errorf("%s: the default branch that returns false is gone", path)
	}
	for _, c := range []domain.TrexContract{domain.TrexContractTricks, domain.TrexContractTrix, domain.TrexContractNone} {
		if domain.TrexCardPenalty(c, king) != 0 {
			t.Errorf("contract %d now penalises a card; the mirror's default branch is wrong", c)
		}
	}
}
