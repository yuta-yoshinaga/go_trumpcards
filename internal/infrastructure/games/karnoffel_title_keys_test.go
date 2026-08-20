package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestKarnoffelTitleKeysMatchTheFrontend guards the two copies of Karnöffel's
// title table against drift.
//
// The Web page names the chosen suit's titled cards from
// `frontend/src/utils/karnoffelRanks.ts` and the CUI now names them from
// `domain.KarnoffelTitleKeys` (#5732). Both are keyed by card value, so a
// one-rank slip (calling the 6 "Kaiser") would render a wrong badge on one side
// only, and every existing test would still pass — each side is internally
// consistent with itself.
func TestKarnoffelTitleKeysMatchTheFrontend(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	path := filepath.Join(root, "frontend", "src", "utils", "karnoffelRanks.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`KARNOFFEL_RANK_KEYS: Readonly<Record<number, string>> = \{([^}]*)\}`).
		FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("karnoffelRanks.ts no longer states KARNOFFEL_RANK_KEYS as an object literal")
	}

	got := map[int]string{}
	for _, line := range strings.Split(m[1], ",") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		value, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		got[value] = strings.Trim(strings.TrimSpace(parts[1]), "'\"")
	}

	// 空振り防止: 片側が読めていないのに一致と読まれるのを防ぐ。
	if len(domain.KarnoffelTitleKeys) != 7 {
		t.Fatalf("domain lists %d titles, want 7", len(domain.KarnoffelTitleKeys))
	}
	if len(got) != len(domain.KarnoffelTitleKeys) {
		t.Fatalf("karnoffelRanks.ts lists %d titles, domain lists %d",
			len(got), len(domain.KarnoffelTitleKeys))
	}
	for value, want := range domain.KarnoffelTitleKeys {
		if got[value] != want {
			t.Errorf("card value %d: frontend says %q, domain says %q", value, got[value], want)
		}
	}
}
