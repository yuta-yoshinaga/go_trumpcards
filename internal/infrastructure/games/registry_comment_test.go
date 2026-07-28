package games_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const registrySrc = "registry.go"

var (
	// A registry entry, e.g. `{Name: "poker", Category: CategoryCasino},`.
	entryRe = regexp.MustCompile(`^\s*\{Name: "([a-z0-9]+)", Category: Category(\w+)\}`)
	// A bucket name used *as a bucket*: adjacent to worker/bucket vocabulary, or
	// the object of a routing verb. Adjacency rather than proximity is deliberate --
	// a window wide enough to span wrapped lines also swept up "a classic Crazy
	// Eights variant" and "a French casino banking game", where the word is a genre,
	// not a bucket. A guard that cries wolf gets deleted by the next person.
	bucketRefRe = regexp.MustCompile(`(?i)\b(extra2|extra3|casino|classic|solo|extra)\b[ \-]*(worker|bucket)` +
		`|(?i)\b(into|off|to|in)\s+(the\s+)?(extra2|extra3|casino|classic|solo|extra)\b` +
		`|(?i)\b(extra2|extra3|casino|classic|solo|extra)\b\s*(ワーカー|バケット)`)
)

// TestRegistryCommentsDoNotNameBuckets keeps per-game comments from naming a
// worker bucket.
//
// The bucket a game lives in is decided by size and moves whenever a worker
// approaches the 1 MB limit: ADR-0027, ADR-0032, ADR-0036 and #4462 have each
// reshuffled it. Every time, the Category field moved and the prose above it did
// not, so by #4462 thirty-five comments named a bucket that disagreed with the
// entry on the very next line -- "bucketed into the casino worker" directly above
// CategoryExtra2.
//
// Rewording them again would not help, because the duplication is the defect:
// the Category field is the only place that can be authoritative, and any copy
// of it in prose is a copy that can rot. So per-entry comments describe the
// *game* and stay silent about the bucket. Which bucket, and when it moved, is
// answered by the field itself plus git history and the ADRs.
//
// The package doc in registry.go already explains that Category is purely a
// size bucket, so nothing is lost by not repeating it 219 times.
func TestRegistryCommentsDoNotNameBuckets(t *testing.T) {
	path := filepath.Join(".", registrySrc)
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(raw), "\n")

	entries := 0
	var offenders []string
	for i, line := range lines {
		m := entryRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		entries++
		game, category := m[1], strings.ToLower(m[2])

		// Walk back over the contiguous comment block directly above the entry.
		var block []string
		for j := i - 1; j >= 0 && strings.HasPrefix(strings.TrimSpace(lines[j]), "//"); j-- {
			block = append(block, strings.TrimSpace(lines[j]))
		}
		if len(block) == 0 {
			continue
		}
		text := strings.Join(block, " ")

		for _, ref := range bucketRefRe.FindAllString(text, -1) {
			offenders = append(offenders, fmt.Sprintf(
				"%s:%d %s says %q (Category=%s); describe the game, not the bucket",
				registrySrc, i+1, game, strings.TrimSpace(ref), category))
		}
	}

	if entries == 0 {
		t.Fatalf("no registry entries parsed from %s -- the entry format changed; update entryRe", path)
	}
	if len(offenders) > 0 {
		t.Errorf("%d per-entry comment(s) name a worker bucket:\n  %s",
			len(offenders), strings.Join(offenders, "\n  "))
	}
}

// TestBucketRefRegexp pins the matcher down in both directions.
//
// The guard above can only fail on comments that exist today, so a branch that
// nothing currently writes -- the Japanese ワーカー/バケット wording -- would sit
// there unexercised and possibly wrong. These cases exercise every branch, and
// just as importantly the two phrasings that must NOT match: a proximity-based
// version of this regex flagged both, which is why it matches adjacency instead.
func TestBucketRefRegexp(t *testing.T) {
	cases := []struct {
		text  string
		match bool
	}{
		{"bucketed into the casino worker purely for size", true},
		{"Classic worker bucket - casino was full", true},
		{"routed to the EXTRA worker", true},
		{"moved off casino in #4462", true},
		{"solo バケットに入れている", true},
		{"casino ワーカーへ移した", true},
		{"Macau is a classic Crazy Eights variant", false},
		{"a French casino banking game - the simplest possible", false},
		{"Barbu is a classic compendium trick-taking game", false},
		{"Category is only a size bucket", false},
	}
	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			if got := bucketRefRe.MatchString(c.text); got != c.match {
				t.Errorf("MatchString(%q) = %v, want %v", c.text, got, c.match)
			}
		})
	}
}
