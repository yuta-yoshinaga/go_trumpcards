//go:build test

package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// workerNames are the seven Cloudflare Workers the games ship to. Listed rather
// than globbed so a worker that loses its wrangler.toml fails here instead of
// silently dropping out of the check.
var workerNames = []string{"casino", "classic", "solo", "extra", "extra2", "extra3", "extra4"}

var observabilityEnabledRe = regexp.MustCompile(`(?m)^\[observability\]\s*$`)

// TestWorkerWranglerEnablesObservability fails on a Worker that ships with its
// logs switched off.
//
// Without `[observability] enabled = true` Cloudflare stores nothing, and the
// observability API answers a seven-day query with zero events. That is not a
// missing nicety: when klondike started returning 503s in production there was
// no log to read, and the cause -- an undo history that grew until the request
// blew the Worker's CPU budget -- had to be found by replaying moves against
// the deployed Worker by hand until one of them failed.
//
// It is easy to lose again, because a new worker is added by copying an
// existing wrangler.toml, and nothing at deploy time complains about the
// absence.
func TestWorkerWranglerEnablesObservability(t *testing.T) {
	root := repoRootFrom(t)

	var missing []string
	checked := 0
	for _, name := range workerNames {
		path := filepath.Join(root, "workers", name, "wrangler.toml")
		body, err := os.ReadFile(path) //nolint:gosec // fixed path under the repo
		if err != nil {
			t.Fatalf("cannot read %s: %v", path, err)
		}
		checked++
		switch {
		case !observabilityEnabledRe.MatchString(string(body)):
			missing = append(missing, name+": no [observability] section")
		case !strings.Contains(string(body), "enabled = true"):
			missing = append(missing, name+": [observability] present but not enabled = true")
		}
	}

	// A loop that read nothing would report success while checking nothing.
	if checked != len(workerNames) {
		t.Fatalf("checked %d wrangler.toml files, expected %d", checked, len(workerNames))
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("Workers shipping with logging off (%d of %d).\n"+
			"Add to workers/<name>/wrangler.toml:\n"+
			"  [observability]\n  enabled = true\n  head_sampling_rate = 1\n"+
			"Without it Cloudflare stores no logs and production can only be\n"+
			"diagnosed by reproducing against it by hand.\n  %s",
			len(missing), len(workerNames), strings.Join(missing, "\n  "))
	}
}

// repoRootFrom walks up from the test's working directory to the module root.
func repoRootFrom(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not find the repository root (no go.mod above the test)")
	return ""
}
