package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// docsPath is the per-worker game table that docs/cloudflare-workers.md carries.
const docsPath = "../../../docs/cloudflare-workers.md"

var (
	docsRowRe = regexp.MustCompile(`(?m)^\| \*\*(\w+)\*\* \| ` + "`" + `cmd/workers/\w+/main\.go` + "`" + ` \| (\d+) \| (.+) \|$`)
	docsKeyRe = regexp.MustCompile("`([a-z0-9]+)`")
)

// TestDocsMatchRegistry asserts that the worker table in docs/cloudflare-workers.md
// lists exactly the games the registry assigns to each bucket.
//
// This guard exists because the table drifted repeatedly while it was curated
// prose: it still described three workers after the fourth shipped, and named
// stale build commands two ADRs later. Rebucketing is routine (ADR-0032,
// ADR-0036) and touches five registration points, so a doc list maintained by
// hand is guaranteed to fall behind. Nothing else in the suite reads this file.
func TestDocsMatchRegistry(t *testing.T) {
	raw, err := os.ReadFile(filepath.Clean(docsPath))
	if err != nil {
		t.Fatalf("read %s: %v", docsPath, err)
	}
	rows := docsRowRe.FindAllStringSubmatch(string(raw), -1)
	if len(rows) == 0 {
		t.Fatalf("no worker rows parsed from %s -- the table format changed; update docsRowRe", docsPath)
	}

	documented := make(map[string][]string, len(rows))
	for _, row := range rows {
		worker, count, cell := row[1], row[2], row[3]
		var keys []string
		for _, m := range docsKeyRe.FindAllStringSubmatch(cell, -1) {
			keys = append(keys, m[1])
		}
		if got := strings.TrimSpace(count); got != itoa(len(keys)) {
			t.Errorf("%s: table says %s games but lists %d keys", worker, got, len(keys))
		}
		documented[worker] = keys
	}

	for _, cat := range []games.Category{
		games.CategoryCasino, games.CategoryClassic, games.CategorySolo,
		games.CategoryExtra, games.CategoryExtra2, games.CategoryExtra3,
	} {
		want := make([]string, 0, 64)
		for _, g := range games.ByCategory(cat) {
			want = append(want, g.Name)
		}
		sort.Strings(want)

		got, ok := documented[cat.String()]
		if !ok {
			t.Errorf("%s: no row in %s", cat, docsPath)
			continue
		}
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s: docs and registry disagree\n  only in docs:     %v\n  only in registry: %v",
				cat, diff(got, want), diff(want, got))
		}
	}

	if len(documented) != 6 {
		t.Errorf("expected 6 worker rows, parsed %d", len(documented))
	}
}

// diff returns the elements of a that are absent from b.
func diff(a, b []string) []string {
	in := make(map[string]bool, len(b))
	for _, s := range b {
		in[s] = true
	}
	var out []string
	for _, s := range a {
		if !in[s] {
			out = append(out, s)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
