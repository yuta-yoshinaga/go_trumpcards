package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestNapWebPayoutMatchesTheDomain guards the web's copy of the contract stakes
// against the values the domain actually pays out.
//
// `frontend/src/utils/napPayout.ts` spells the Nap asymmetry (10 to make, 5
// apiece to beat) as literals so the page can label its bid buttons. Change
// makeValue/failValue in `internal/domain/Nap.go` and the buttons would keep
// promising the old stake with nothing to notice (#5651) -- the same drift the
// Baccarat payout panel had in #5497.
func TestNapWebPayoutMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "utils", "napPayout.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	// The Nap branch is the only literal pair in the file.
	m := regexp.MustCompile(`return \{ make: (\d+), fail: (\d+) \};`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("napPayout.ts no longer states the Nap stake as a literal pair")
	}
	wantMake, wantFail := domain.NapBidPayout(domain.NapBidNap)
	if m[1] != strconv.Itoa(wantMake) || m[2] != strconv.Itoa(wantFail) {
		t.Errorf("napPayout.ts Nap = %s/%s, want %d/%d", m[1], m[2], wantMake, wantFail)
	}

	// The numbered contracts stake their own trick count both ways; if that ever
	// stops being true the web's `{ make: contract, fail: contract }` is wrong.
	for _, bid := range []domain.NapBid{domain.NapBidTwo, domain.NapBidThree, domain.NapBidFour} {
		gotMake, gotFail := domain.NapBidPayout(bid)
		if gotMake != int(bid) || gotFail != int(bid) {
			t.Errorf("NapBidPayout(%v) = %d/%d, but napPayout.ts assumes %d/%d",
				bid, gotMake, gotFail, int(bid), int(bid))
		}
	}
}
