package games_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// repoRoot is the module root relative to this test's working directory
// (the package dir internal/infrastructure/games).
const repoRoot = "../../.."

// readDocNumber reads the file at path (relative to repoRoot), applies re —
// which must capture exactly one numeric group — and returns the captured
// integer. A missing match fails loudly so a reworded sentence surfaces here
// (update the regex) rather than silently skipping the assertion.
func readDocNumber(t *testing.T, path string, re *regexp.Regexp) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := re.FindSubmatch(data)
	if m == nil {
		t.Fatalf("%s: pattern %q not found — did the wording change? update the regex in docs_consistency_test.go", path, re)
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil {
		t.Fatalf("%s: matched %q is not an integer: %v", path, m[1], err)
	}
	return n
}

// TestDocGameCountsMatchRegistry guards the hand-maintained game-count numbers
// in README.md and CLAUDE.md against the registry single source of truth.
// These prose counts drift on every new game (see issue #2474, where README
// still said 126 after the registry reached 132); this test fails CI the
// moment they fall out of sync instead of waiting for a manual drift sweep.
func TestDocGameCountsMatchRegistry(t *testing.T) {
	total := len(games.All())

	cases := []struct {
		name string
		path string
		re   *regexp.Regexp
	}{
		{"README.md Features prose", "README.md", regexp.MustCompile(`実装した(\d+)種類`)},
		{"CLAUDE.md intro line", "CLAUDE.md", regexp.MustCompile(`Go implementations of (\d+) trump card game`)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := readDocNumber(t, c.path, c.re); got != total {
				t.Errorf("%s says %d games, registry has %d — update the doc (or the regex if the wording moved)", c.name, got, total)
			}
		})
	}
}

// archEndpointRe captures the game name from every `POST /<name>/exec`
// reference in docs/architecture.md. Game names are lowercase alphanumeric
// (e.g. blackjack, omahahilo, deucetoseven), matching the registry Name field.
var archEndpointRe = regexp.MustCompile(`POST /([a-z0-9]+)/exec`)

// TestArchitectureDocEndpointsMatchRegistry guards the hand-maintained Web API
// endpoint list in docs/architecture.md against the registry single source of
// truth. That list (and its spelled-out count) drifted badly — it fell 21
// games behind while still claiming "One hundred forty-eight endpoints" (see
// issue #2525) — because, unlike README.md/CLAUDE.md, no test covered it. This
// asserts a strict 1:1 mapping between registered games and the `POST
// /<name>/exec` entries in the doc, failing CI the moment one is added or
// removed without updating the doc.
func TestArchitectureDocEndpointsMatchRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/architecture.md"))
	if err != nil {
		t.Fatalf("read docs/architecture.md: %v", err)
	}

	matches := archEndpointRe.FindAllSubmatch(data, -1)
	docNames := make(map[string]bool)
	for _, m := range matches {
		docNames[string(m[1])] = true
	}

	registryNames := make(map[string]bool, len(games.All()))
	for _, g := range games.All() {
		registryNames[g.Name] = true
	}

	// Catch duplicate entries: the maps above dedupe, so a game listed twice
	// would still pass the 1:1 checks below. Comparing the raw match count to
	// the (deduped) registry size surfaces a doubled endpoint.
	if len(matches) != len(registryNames) {
		t.Errorf("docs/architecture.md has %d POST /<name>/exec entries but registry has %d games — check for duplicate or stray endpoint lines", len(matches), len(registryNames))
	}

	// Every registered game must have a POST /<name>/exec entry in the doc.
	for name := range registryNames {
		if !docNames[name] {
			t.Errorf("docs/architecture.md is missing endpoint POST /%s/exec for registered game %q", name, name)
		}
	}
	// No orphan endpoints: every documented endpoint must map to a real game.
	for name := range docNames {
		if !registryNames[name] {
			t.Errorf("docs/architecture.md documents endpoint POST /%s/exec with no matching game in the registry", name)
		}
	}
}

// archEndpointCountRe captures the endpoint totals stated in
// docs/architecture.md -- the summary bullet and the table's own header line.
var archEndpointCountRe = regexp.MustCompile(`(?:\*\*Web API\*\*: |-- \*\*)(\d+)(?:\*\* in total| endpoints, one per game)`)

// TestArchitectureDocEndpointCountMatchesRegistry guards the endpoint totals
// written in prose.
//
// Until #4470 that number was spelled out in words ("Two hundred and thirty-two
// endpoints") inside a single paragraph that also held every endpoint, and
// nothing checked it: TestArchitectureDocEndpointsMatchRegistry only counts
// `POST /<name>/exec` occurrences, so the prose could say anything. Every game
// added this session updated it by hand on trust.
func TestArchitectureDocEndpointCountMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/architecture.md"))
	if err != nil {
		t.Fatalf("read docs/architecture.md: %v", err)
	}

	matches := archEndpointCountRe.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		t.Fatal("docs/architecture.md states no endpoint count -- update the regex if the wording moved")
	}

	want := len(games.All())
	for _, m := range matches {
		got, err := strconv.Atoi(string(m[1]))
		if err != nil {
			t.Fatalf("unparsable endpoint count %q: %v", m[1], err)
		}
		if got != want {
			t.Errorf("docs/architecture.md states %d endpoints, registry has %d — update the doc", got, want)
		}
	}
}

// TestPerGameManualsMatchRegistry asserts a strict 1:1 mapping between
// registered games and the per-game manuals under docs/manual/{cui,web}. It
// catches both a missing manual for a freshly added game and an orphan manual
// left behind by a rename or removal. Template files live at
// docs/manual/{cui,web}_template.md (outside these dirs) and are not counted.
func TestPerGameManualsMatchRegistry(t *testing.T) {
	all := games.All()
	names := make(map[string]bool, len(all))
	for _, g := range all {
		names[g.Name] = true
	}

	for _, dir := range []string{"docs/manual/cui", "docs/manual/web"} {
		t.Run(dir, func(t *testing.T) {
			// Every registered game must have a manual in this dir.
			for _, g := range all {
				p := filepath.Join(repoRoot, dir, g.Name+".md")
				if _, err := os.Stat(p); err != nil {
					t.Errorf("missing manual %s/%s.md for registered game %q", dir, g.Name, g.Name)
				}
			}
			// No orphan manuals: every <name>.md must map to a registered game.
			entries, err := os.ReadDir(filepath.Join(repoRoot, dir))
			if err != nil {
				t.Fatalf("read dir %s: %v", dir, err)
			}
			for _, e := range entries {
				name, ok := strings.CutSuffix(e.Name(), ".md")
				if !ok {
					t.Errorf("%s/%s is not a .md manual file", dir, e.Name())
					continue
				}
				if !names[name] {
					t.Errorf("orphan manual %s/%s.md has no matching game in the registry", dir, name)
				}
			}
		})
	}
}

// bucketEnumRe matches a brace enumeration such as
// `{casino,classic,solo,extra,extra2,extra3}` as used in build commands and
// path globs throughout the docs.
var bucketEnumRe = regexp.MustCompile(`\{([a-z0-9]+(?:,[a-z0-9]+)+)\}`)

// workerPathRe matches a per-worker entry point path, one per table row in the
// worker list.
var workerPathRe = regexp.MustCompile(`cmd/workers/([a-z0-9]+)/main\.go`)

// docsExemptFromBucketEnum are the docs whose bucket lists are deliberately
// frozen. ADRs record what was true when the decision was made -- ADR-0031
// names `{casino,classic,solo}` because there were three workers then, and
// rewriting that to six would falsify the record. Nothing else belongs here:
// an exemption for a live doc is the drift this guard exists to catch.
var docsExemptFromBucketEnum = []string{"docs/adr/"}

// TestDocsEnumerateEveryWorkerBucket asserts that a doc which enumerates worker
// buckets enumerates ALL of them.
//
// Two rounds of #4474 missed sites because I searched for the *wording* I
// remembered writing ("four Cloudflare Workers", "overflow bucket") instead of
// for the *enumeration*. A list spelling out `{casino,classic,solo}` contains
// neither the old count nor the new one, so no phrase search could ever reach
// it: docs/architecture.md still described three workers and 93 of 233 games
// two ADRs after that stopped being true, and docs/new-game-checklist.md:15
// still named four. Both are mechanical to check, so check them mechanically.
//
// The rule is deliberately narrow -- it fires only on lists that already name
// at least two buckets, so `docs/manual/{cui,web}` and friends are untouched.
func TestDocsEnumerateEveryWorkerBucket(t *testing.T) {
	all := make(map[string]bool, len(games.AllCategories()))
	var want []string
	for _, c := range games.AllCategories() {
		all[c.String()] = true
		want = append(want, c.String())
	}
	sort.Strings(want)

	// requireComplete reports whether names -- a set of bucket-ish tokens found
	// in one place in one file -- is either bucket-free, or the full set.
	requireComplete := func(t *testing.T, path, context string, names []string) {
		t.Helper()
		var found []string
		for _, n := range names {
			if all[n] {
				found = append(found, n)
			}
		}
		if len(found) < 2 {
			return // not a bucket enumeration
		}
		sort.Strings(found)
		found = slices.Compact(found)
		if !slices.Equal(found, want) {
			t.Errorf("%s: %s enumerates buckets %v, registry has %v -- a partial list is how this doc drifted before",
				path, context, found, want)
		}
	}

	// Walk docs/ rather than globbing it: `docs/*.md` does not cross a `/`, so
	// it would silently skip every subdirectory -- including docs/adr, which
	// would leave the exemption below guarding nothing while reading as though
	// it were load-bearing.
	docs := []string{filepath.Join(repoRoot, "CLAUDE.md"), filepath.Join(repoRoot, "internal/CLAUDE.md")}
	if err := filepath.WalkDir(filepath.Join(repoRoot, "docs"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			docs = append(docs, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk docs: %v", err)
	}

	for _, doc := range docs {
		rel, err := filepath.Rel(repoRoot, doc)
		if err != nil {
			t.Fatalf("rel %s: %v", doc, err)
		}
		rel = filepath.ToSlash(rel)
		if slices.ContainsFunc(docsExemptFromBucketEnum, func(p string) bool { return strings.HasPrefix(rel, p) }) {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(doc))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}

		for _, m := range bucketEnumRe.FindAllStringSubmatch(string(data), -1) {
			requireComplete(t, rel, "brace list "+m[0], strings.Split(m[1], ","))
		}

		// A per-worker table names one entry point per row, so the set of
		// referenced entry points is itself an enumeration.
		var paths []string
		for _, m := range workerPathRe.FindAllStringSubmatch(string(data), -1) {
			paths = append(paths, m[1])
		}
		requireComplete(t, rel, "cmd/workers/* references", paths)
	}
}

// openapiPathRe matches one `  /<game>/exec:` key in the OpenAPI spec.
var openapiPathRe = regexp.MustCompile(`(?m)^  /([a-z0-9]+)/exec:`)

// TestOpenAPIMatchesRegistry asserts that api/openapi.yaml documents exactly
// the games the registry holds -- one POST /<game>/exec per game, no more.
//
// This is the last per-game file that nothing checked. docs/cloudflare-workers.md,
// docs/architecture.md, docs/manual/{cui,web} and frontend/src/api/gameExec.ts
// all have guards; openapi.yaml had only a line in the new-game checklist, and
// it drifted by four games (braid, pontoon, settemezzo, niuniu) before anyone
// looked. A rule that is only written down is a rule that gets skipped -- that
// is the whole reason the other guards exist.
//
// api/openapi.yaml is CRLF. The regex tolerates that because `$` in Go's
// multiline mode stops before the \r, but anything that rewrites the file must
// preserve the line endings or the diff becomes every line.
func TestOpenAPIMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "api/openapi.yaml"))
	if err != nil {
		t.Fatalf("read api/openapi.yaml: %v", err)
	}

	documented := map[string]bool{}
	for _, m := range openapiPathRe.FindAllSubmatch(data, -1) {
		documented[string(m[1])] = true
	}
	if len(documented) == 0 {
		t.Fatal("no /<game>/exec paths parsed from api/openapi.yaml -- the format changed; update openapiPathRe")
	}

	registered := map[string]bool{}
	for _, g := range games.All() {
		registered[g.Name] = true
	}

	var missing, orphaned []string
	for name := range registered {
		if !documented[name] {
			missing = append(missing, name)
		}
	}
	for name := range documented {
		if !registered[name] {
			orphaned = append(orphaned, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(orphaned)

	if len(missing) > 0 {
		t.Errorf("registered games with no OpenAPI path: %v -- add POST /<game>/exec to api/openapi.yaml", missing)
	}
	if len(orphaned) > 0 {
		t.Errorf("OpenAPI paths for games that are not registered: %v -- a rename or removal left them behind", orphaned)
	}
}

// openapiRefRe matches a `$ref: '#/components/schemas/X'` pointer, and
// openapiSchemaRe a schema definition at the fixed four-space indent the file
// uses under components.schemas.
var (
	openapiRefRe = regexp.MustCompile(`\$ref: '#/components/schemas/([A-Za-z0-9]+)'`)
	// The trailing \r? is load-bearing: api/openapi.yaml is CRLF, so `$` sits
	// after the carriage return and an anchored pattern matches nothing --
	// which reads as "every reference is dangling" rather than as a broken
	// regex.
	openapiSchemaRe = regexp.MustCompile(`(?m)^    ([A-Za-z0-9]+):\r?$`)
)

// TestOpenAPIHasNoDanglingSchemaRefs asserts that every $ref points at a schema
// that exists.
//
// Two of these were already in the file. `SirTommyHint` was referenced by the
// Sir Tommy response and never defined; `ErrorResponse` I invented myself in
// the Bura change -- I wrote the 400 branch from memory instead of copying the
// convention, which is that a 400 carries the game's own response payload with
// a message. A spec that points at a schema which is not there generates
// broken clients, and neither reference cost anything to add unnoticed.
func TestOpenAPIHasNoDanglingSchemaRefs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "api/openapi.yaml"))
	if err != nil {
		t.Fatalf("read api/openapi.yaml: %v", err)
	}
	text := string(data)

	defined := map[string]bool{}
	// Only the components.schemas block defines schemas; the four-space indent
	// is unique to it in this file, but confirm the section exists so a
	// restructure fails loudly instead of silently matching nothing.
	if !strings.Contains(text, "\n  schemas:\n") && !strings.Contains(text, "\r\n  schemas:\r\n") {
		t.Fatal("no components.schemas block found -- the file structure changed")
	}
	for _, m := range openapiSchemaRe.FindAllStringSubmatch(text, -1) {
		defined[m[1]] = true
	}

	var dangling []string
	seen := map[string]bool{}
	for _, m := range openapiRefRe.FindAllStringSubmatch(text, -1) {
		if !defined[m[1]] && !seen[m[1]] {
			seen[m[1]] = true
			dangling = append(dangling, m[1])
		}
	}
	sort.Strings(dangling)

	if len(dangling) > 0 {
		t.Errorf("api/openapi.yaml references schemas that are not defined: %v", dangling)
	}
}

// openapiExecKeyRe matches the start of one `  /<game>/exec:` path key. The
// blocks are cut by splitting on these rather than by matching a block whose
// terminator is the NEXT key: RE2 has no lookahead, so a block pattern
// consumes the following key and silently skips every other path. That is not
// hypothetical -- the first version of this guard checked 72 of 234 paths and
// reported one of the three real mismatches.
var openapiExecKeyRe = regexp.MustCompile(`(?m)^  /([a-z0-9]+)/exec:\r?$`)

// openapiStatusRefRe captures a status code and the schema its response body
// references, e.g. ('200', 'BuraResponse').
var openapiStatusRefRe = regexp.MustCompile(`'(\d{3})':(?s:.*?)\$ref: '#/components/schemas/([A-Za-z0-9]+)'`)

// TestOpenAPIErrorResponseMatchesTheSuccessSchema asserts that a path's 400
// documents the same schema as its 200.
//
// Every endpoint here returns the game's own payload on both branches -- an
// error arrives as a normal response carrying a `message`, not as a separate
// error type. So the two refs must agree, and when they do not the spec
// describes some other game's shape to anyone generating a client.
//
// This exists because I broke exactly that. Fixing the invented `ErrorResponse`
// ref, I replaced "the first remaining occurrence" once per game while
// iterating the games in a different order than they appear in the file, so
// three of the five 400s landed on a sibling's schema. Two were right by
// coincidence, which is what made it survive a read-through.
//
// TestOpenAPIHasNoDanglingSchemaRefs cannot catch this: every one of those
// names is defined, just not the right one for its path.
func TestOpenAPIErrorResponseMatchesTheSuccessSchema(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "api/openapi.yaml"))
	if err != nil {
		t.Fatalf("read api/openapi.yaml: %v", err)
	}

	text := string(data)
	locs := openapiExecKeyRe.FindAllStringSubmatchIndex(text, -1)
	if len(locs) == 0 {
		t.Fatal("no /<game>/exec keys parsed -- the file structure changed; update openapiExecKeyRe")
	}

	checked := 0
	for i, loc := range locs {
		game := text[loc[2]:loc[3]]
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		body := text[loc[1]:end]
		refs := map[string]string{}
		for _, m := range openapiStatusRefRe.FindAllStringSubmatch(body, -1) {
			if _, seen := refs[m[1]]; !seen {
				refs[m[1]] = m[2]
			}
		}
		ok, bad := refs["200"], refs["400"]
		if ok == "" || bad == "" {
			continue // a path documenting only one branch is not this test's business
		}
		checked++
		if ok != bad {
			t.Errorf("/%s/exec: 200 documents %s but 400 documents %s -- the 400 must carry the same game's payload",
				game, ok, bad)
		}
	}

	if checked == 0 {
		t.Fatal("no path documented both a 200 and a 400 -- the parse found nothing to check")
	}
	t.Logf("checked %d paths", checked)
}
