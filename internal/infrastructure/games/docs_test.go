package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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
		if got := strings.TrimSpace(count); got != strconv.Itoa(len(keys)) {
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

	// Derived from the registry, not hardcoded: this assertion read `!= 6` when
	// ADR-0037 added the seventh bucket, so the guard failed on a correct table
	// rather than on a stale one.
	if want := len(games.AllCategories()); len(documented) != want {
		t.Errorf("expected %d worker rows, parsed %d", want, len(documented))
	}
}

// gameExecPath holds the frontend's game -> Worker URL map.
const gameExecPath = "../../../frontend/src/api/gameExec.ts"

var gameExecRe = regexp.MustCompile(`(?m)^\s+([a-z0-9]+): WORKER_([A-Z0-9]+),`)

// TestGameExecMatchesRegistry asserts that frontend/src/api/gameExec.ts routes every
// game to the worker its registry Category names.
//
// This map is the one touchpoint where a mistake is invisible everywhere else: Go
// still builds, the worker still builds, every test still passes, and the game simply
// 404s in production because the browser asks a worker whose binary does not contain
// it. Nothing but the move script kept it in step with registry.go, and a script only
// helps the moves that actually use it.
func TestGameExecMatchesRegistry(t *testing.T) {
	raw, err := os.ReadFile(filepath.Clean(gameExecPath))
	if err != nil {
		t.Fatalf("read %s: %v", gameExecPath, err)
	}
	matches := gameExecRe.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatalf("no workerUrl entries parsed from %s -- the map format changed; update gameExecRe", gameExecPath)
	}

	routed := make(map[string]string, len(matches))
	for _, m := range matches {
		routed[m[1]] = strings.ToLower(m[2])
	}

	for _, g := range games.All() {
		worker, ok := routed[g.Name]
		if !ok {
			t.Errorf("%s: missing from workerUrl in %s", g.Name, gameExecPath)
			continue
		}
		if worker != g.Category.String() {
			t.Errorf("%s: routed to %q but registry says %q", g.Name, worker, g.Category)
		}
	}
	for name := range routed {
		if !registered(name) {
			t.Errorf("%s: routed in %s but absent from the registry", name, gameExecPath)
		}
	}
}

// registered reports whether name is a game in the registry.
func registered(name string) bool {
	for _, g := range games.All() {
		if g.Name == name {
			return true
		}
	}
	return false
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
