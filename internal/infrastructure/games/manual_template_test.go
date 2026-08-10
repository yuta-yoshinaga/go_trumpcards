package games_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// TestPerGameManualsFollowTemplate guards the structure of every per-game
// manual against docs/manual/{cui,web}_template.md. Its sibling
// TestPerGameManualsMatchRegistry (docs_consistency_test.go) checks only that a
// manual exists; nothing checked what was inside one, and 380 of the 528
// manuals had drifted by 2026-08: 110 CUI command tables never mentioned
// `help`, 106 Web manuals omitted `go run ./cmd/server`, and 12 CUI manuals
// still told the reader to run `go run ./cmd/cli <game>` -- a binary this repo
// does not have. These files are not internal notes:
// frontend/src/constants/{cuiManualTexts,manualTexts}.ts glob-import them into
// the Web GUI, so a stale launch line ships to players.
//
// Three things this test deliberately does not do:
//
//   - It does not parse the Mermaid. frontend/scripts/check-mermaid.mjs already
//     runs the real mermaid.parse() over every block in the repo. Adding a
//     second, weaker parser here would just disagree with it eventually.
//   - It does not require 遊び方のコツ. Both templates mark that section
//     `<!-- 任意セクション -->`.
//   - It has no allowlist. An exemption here is precisely the drift this guard
//     exists to catch -- see docsExemptFromBucketEnum's own warning.
//
// scripts/audit-manual-template.mjs asserts the same rules and prints a
// worklist grouped by issue class. Keep the two in sync.
func TestPerGameManualsFollowTemplate(t *testing.T) {
	all := games.All()
	nav := loadNavLabels(t)
	if len(nav) < len(all) {
		t.Fatalf("frontend/src/i18n/locales/ja/common.json yielded %d nav labels for %d games — "+
			"the `nav` object moved or was renamed; this test cannot check H1s without it", len(nav), len(all))
	}

	checked := 0
	for _, spec := range manualSpecs {
		for _, g := range all {
			rel := filepath.Join(spec.dir, g.Name+".md")
			data, err := os.ReadFile(filepath.Join(repoRoot, rel))
			if err != nil {
				// A missing manual is TestPerGameManualsMatchRegistry's report
				// to make; failing twice for one cause helps nobody.
				continue
			}
			checked++
			t.Run(rel, func(t *testing.T) {
				checkManual(t, spec, g.Name, nav[g.Name], string(data), rel)
			})
		}
	}
	if checked != 2*len(all) {
		t.Fatalf("checked %d manuals but the registry has %d games across two directories — "+
			"the walk broke rather than the manuals being clean; fix it in manual_template_test.go",
			checked, 2*len(all))
	}
}

// manualSpec is one directory's contract, in the order the template lays the
// sections out.
type manualSpec struct {
	dir      string
	kind     string // the version marker inside the H1 parenthesis
	sections []string
}

var manualSpecs = []manualSpec{
	{
		dir:      "docs/manual/cui",
		kind:     "CUI",
		sections: []string{"ゲーム概要", "起動方法", "ルール", "ゲームの流れ", "コマンド一覧", "画面の見方"},
	},
	{
		dir:      "docs/manual/web",
		kind:     "Web",
		sections: []string{"ゲーム概要", "起動方法", "ルール", "ゲームの流れ", "画面の操作方法", "画面構成"},
	},
}

// manualFenceRe matches a fence opening or closing a code block.
var manualFenceRe = regexp.MustCompile("^\\s*(?:`{3,}|~{3,})")

// manualCommandRowRe builds the check for one mandatory command-table row. The
// command may carry arguments -- spider documents `reset [1\|2\|4]` -- so it
// matches the token rather than requiring a closing backtick right after it.
func manualCommandRowRe(cmd string) *regexp.Regexp {
	return regexp.MustCompile(`\|\s*` + "`" + regexp.QuoteMeta(cmd) + `\b`)
}

// manualAPIPathRe matches an HTTP route written outside a code fence. The Web
// template says API specs belong in api/openapi.yaml, not in a manual.
var manualAPIPathRe = regexp.MustCompile(`^\s*(?:GET|POST|PUT|PATCH|DELETE)\s+/`)

// loadNavLabels reads the Japanese navigation label for every game. This is the
// same string the Web GUI shows in its navigation, so a manual's H1 and the nav
// entry for one game cannot disagree.
func loadNavLabels(t *testing.T) map[string]string {
	t.Helper()
	path := filepath.Join(repoRoot, "frontend/src/i18n/locales/ja/common.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var parsed struct {
		Nav map[string]string `json:"nav"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return parsed.Nav
}

// manualHeadings splits a manual into its H1s and its ordered H2s, ignoring
// anything inside a fenced code block.
//
// The fence skip is load-bearing, not defensive: a `# 英語表示:` comment inside
// a ```sh block is not a heading, and counting it as one produced 13 phantom
// "duplicate H1" reports the first time these manuals were measured.
func manualHeadings(src string) (h1 []string, h2 []string) {
	inFence := false
	for line := range strings.SplitSeq(src, "\n") {
		if manualFenceRe.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		switch {
		case strings.HasPrefix(line, "## "):
			h2 = append(h2, strings.TrimSpace(line[3:]))
		case strings.HasPrefix(line, "# "):
			h1 = append(h1, strings.TrimSpace(line))
		}
	}
	return h1, h2
}

// manualSection returns the body of one H2 section, fences included, and
// whether it is present. The boundary is located on the fence-aware heading
// scan: searching raw lines ended four manuals' 起動方法 at a `# または`
// comment inside their ```sh block, hiding the `go run ./cmd/server` that
// followed it.
func manualSection(src, name string) (string, bool) {
	lines := strings.Split(src, "\n")
	inFence, start := false, -1
	for i, line := range lines {
		if manualFenceRe.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence || !strings.HasPrefix(line, "## ") {
			continue
		}
		if start >= 0 {
			return strings.Join(lines[start+1:i], "\n"), true
		}
		if strings.TrimSpace(line[3:]) == name {
			start = i
		}
	}
	if start < 0 {
		return "", false
	}
	return strings.Join(lines[start+1:], "\n"), true
}

// checkManual reports every way one manual departs from its template. All
// failures are t.Errorf so a single run lists every problem in the file.
func checkManual(t *testing.T, spec manualSpec, game, navLabel, src, rel string) {
	t.Helper()
	h1, h2 := manualHeadings(src)

	wantTitle := fmt.Sprintf("# %s（%s版）遊び方", navLabel, spec.kind)
	switch {
	case len(h1) == 0:
		t.Errorf("%s: no H1. Want %q per docs/manual/%s_template.md", rel, wantTitle, strings.ToLower(spec.kind))
	case len(h1) > 1:
		t.Errorf("%s: %d H1 headings, want exactly one (%q)", rel, len(h1), wantTitle)
	case h1[0] != wantTitle:
		t.Errorf("%s: H1 is %q, want %q. The game name comes from nav.%s in "+
			"frontend/src/i18n/locales/ja/common.json — the label the Web GUI's navigation shows.",
			rel, h1[0], wantTitle, game)
	}

	// Presence and order in one comparison: filter the document's H2 sequence
	// down to the required names and compare it to the template's order.
	var got []string
	for _, h := range h2 {
		if slices.Contains(spec.sections, h) && !slices.Contains(got, h) {
			got = append(got, h)
		}
	}
	if !slices.Equal(got, spec.sections) {
		t.Errorf("%s: required sections are %v, want exactly %v in that order (docs/manual/%s_template.md)",
			rel, got, spec.sections, strings.ToLower(spec.kind))
	}
	// Counted by exact equality, not by substring. A NUL-delimited
	// strings.Count matched any permitted heading that merely starts with a
	// required name -- `## ルール一覧` beside `## ルール` would have been
	// reported as a duplicate. It also diverged from the JS twin, which counts
	// with a plain equality filter.
	for _, s := range spec.sections {
		n := 0
		for _, h := range h2 {
			if h == s {
				n++
			}
		}
		if n > 1 {
			t.Errorf("%s: section %q appears %d times, want once", rel, s, n)
		}
	}

	if flow, ok := manualSection(src, "ゲームの流れ"); ok {
		if !hasMermaidFlowchart(flow) {
			t.Errorf("%s: ゲームの流れ has no ```mermaid block starting with `flowchart`. "+
				"Both templates mark the flowchart mandatory. (Diagram validity is "+
				"frontend/scripts/check-mermaid.mjs's job — do not add a parser here.)", rel)
		}
		// Exactly one. Twelve manuals briefly carried two: they already had a
		// hand-written diagram under ルール, an authoring pass added a second
		// derived one here, and a later pass then rewrote whichever came first
		// in the file. A reader cannot tell which of two diagrams is current.
		if n := strings.Count(flow, "```mermaid"); n > 1 {
			t.Errorf("%s: ゲームの流れ has %d ```mermaid blocks, want exactly one — "+
				"a second diagram leaves the reader guessing which one is current", rel, n)
		}
	}

	launch, _ := manualSection(src, "起動方法")
	if spec.kind == "CUI" {
		want := "go run ./cmd/trumpcards " + game
		if !strings.Contains(launch, want) {
			t.Errorf("%s: 起動方法 does not document %q", rel, want)
		}
		if strings.Contains(launch, "./cmd/cli") {
			t.Errorf("%s: 起動方法 tells the reader to run `go run ./cmd/cli %s`, but cmd/ holds "+
				"only server, trumpcards and workers — use `go run ./cmd/trumpcards %s`", rel, game, game)
		}
		cmds, _ := manualSection(src, "コマンド一覧")
		if !regexp.MustCompile(`\|\s*コマンド\s*\|\s*短縮形\s*\|\s*説明\s*\|`).MatchString(cmds) {
			t.Errorf("%s: コマンド一覧 has no `| コマンド | 短縮形 | 説明 |` table header", rel)
		}
		for _, c := range []string{"reset", "quit", "help"} {
			if !manualCommandRowRe(c).MatchString(cmds) {
				t.Errorf("%s: コマンド一覧 has no `%s` row. Every game accepts it — execCuiCommand in "+
					"internal/adapter/controller/cui_controller_helper.go handles it before the game does.", rel, c)
			}
		}
	} else {
		for _, want := range []string{"go run ./cmd/trumpcards web", "go run ./cmd/server"} {
			if !strings.Contains(launch, want) {
				t.Errorf("%s: 起動方法 does not document %q", rel, want)
			}
		}
		if line, ok := firstAPIPathOutsideFence(src); ok {
			t.Errorf("%s: documents the HTTP API (%q). The Web template keeps API specs in "+
				"api/openapi.yaml so they cannot drift from the served routes.", rel, line)
		}
	}

	for _, leftover := range []string{"＜ゲーム名＞", "<!-- テンプレート:", "<!-- API仕様は"} {
		if strings.Contains(src, leftover) {
			t.Errorf("%s: still contains the template scaffolding %q", rel, leftover)
		}
	}
}

// hasMermaidFlowchart reports whether body holds a ```mermaid block whose first
// non-blank line starts with `flowchart`.
func hasMermaidFlowchart(body string) bool {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "```mermaid" {
			continue
		}
		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) == "" {
				continue
			}
			return strings.HasPrefix(strings.TrimSpace(next), "flowchart")
		}
	}
	return false
}

// firstAPIPathOutsideFence finds an HTTP route written as prose. Routes inside
// code fences are fine — a manual may show a curl example — and comments are
// skipped so the template's own "keep API specs out" note is not a hit.
func firstAPIPathOutsideFence(src string) (string, bool) {
	inFence := false
	for line := range strings.SplitSeq(src, "\n") {
		if manualFenceRe.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence || strings.Contains(line, "<!--") {
			continue
		}
		if manualAPIPathRe.MatchString(line) {
			return strings.TrimSpace(line), true
		}
	}
	return "", false
}

// conformingManual is the minimum CUI manual that satisfies every rule above.
// The mutations in TestManualTemplateGuardCatchesEachViolation each break
// exactly one of them, so the guard is checked in both directions: it must stay
// silent on a good file and must fire on each specific defect. A guard verified
// only against the repository's own (now clean) manuals would pass just as
// happily with its checks deleted.
const conformingManual = "# テスト（CUI版）遊び方\n" + `
## ゲーム概要

テスト用。

## 起動方法

` + "```sh" + `
go run ./cmd/trumpcards testgame
# または
go run ./cmd/trumpcards --lang en testgame
` + "```" + `

## ルール

無し。

## ゲームの流れ

` + "```mermaid" + `
flowchart TD
    A[開始] --> B[終了]
` + "```" + `

## コマンド一覧

| コマンド | 短縮形 | 説明 |
|----------|--------|------|
| ` + "`reset`" + ` | ` + "`r`" + ` | 新しいゲームを開始 |
| ` + "`quit`" + ` | ` + "`q`" + ` | ゲーム終了 |
| ` + "`help`" + ` | ` + "`?`" + ` | コマンド一覧を表示 |

## 画面の見方

` + "```" + `
==========
テスト
==========
# これは見出しではない
` + "```" + `
`

// cuiSpec is the CUI contract, looked up rather than duplicated so the control
// tests cannot drift from the rules the real run enforces.
var cuiSpec = manualSpecs[0]

func TestManualTemplateGuardAcceptsAConformingManual(t *testing.T) {
	fake := &testing.T{}
	checkManual(fake, cuiSpec, "testgame", "テスト", conformingManual, "docs/manual/cui/testgame.md")
	if fake.Failed() {
		t.Error("the guard rejected a manual that follows the template — it would fire on correct work")
	}
}

func TestManualTemplateGuardCatchesEachViolation(t *testing.T) {
	cases := []struct {
		name string
		mut  func(string) string
	}{
		{"wrong H1", func(s string) string {
			return strings.Replace(s, "# テスト（CUI版）遊び方", "# テスト (American Toad) — CUI マニュアル", 1)
		}},
		{"second H1", func(s string) string { return s + "\n# もう一つ\n" }},
		{"missing section", func(s string) string { return strings.Replace(s, "## ルール", "## 規則", 1) }},
		{"sections out of order", func(s string) string {
			return strings.Replace(s, "## ルール", "## 画面の見方", 1)
		}},
		{"duplicate section", func(s string) string { return s + "\n## ルール\n\nふたつめ。\n" }},
		{"mermaid replaced by a state diagram", func(s string) string {
			return strings.Replace(s, "flowchart TD", "stateDiagram-v2", 1)
		}},
		{"a second diagram in ゲームの流れ", func(s string) string {
			return strings.Replace(s, "```mermaid\nflowchart TD\n    A[開始] --> B[終了]\n```",
				"```mermaid\nflowchart TD\n    A[開始] --> B[終了]\n```\n\n```mermaid\nflowchart TD\n    X[古い図] --> Y[終了]\n```", 1)
		}},
		{"mermaid block removed", func(s string) string {
			return strings.Replace(s, "```mermaid", "```text", 1)
		}},
		{"launch command missing", func(s string) string {
			return strings.Replace(s, "go run ./cmd/trumpcards testgame", "go run ./cmd/x testgame", 1)
		}},
		{"launch command points at the deleted cmd/cli", func(s string) string {
			return strings.Replace(s, "go run ./cmd/trumpcards testgame", "go run ./cmd/cli testgame", 1)
		}},
		{"command table lost its 短縮形 column", func(s string) string {
			return strings.Replace(s, "| コマンド | 短縮形 | 説明 |", "| コマンド | 説明 |", 1)
		}},
		{"help row missing", func(s string) string {
			return strings.Replace(s, "| `help` | `?` | コマンド一覧を表示 |", "", 1)
		}},
		{"reset row missing", func(s string) string {
			return strings.Replace(s, "| `reset` | `r` | 新しいゲームを開始 |", "", 1)
		}},
		{"template scaffolding left behind", func(s string) string {
			return strings.Replace(s, "テスト用。", "＜ゲーム名＞ の説明", 1)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fake := &testing.T{}
			checkManual(fake, cuiSpec, "testgame", "テスト", c.mut(conformingManual), "docs/manual/cui/testgame.md")
			if !fake.Failed() {
				t.Errorf("the guard stayed silent on %q — that defect would ship unnoticed", c.name)
			}
		})
	}
}

// The Web contract adds two launch commands and forbids HTTP routes in prose.
func TestManualTemplateGuardChecksTheWebSpecifics(t *testing.T) {
	webSpec := manualSpecs[1]
	base := strings.NewReplacer(
		"（CUI版）", "（Web版）",
		"## コマンド一覧", "## 画面の操作方法",
		"## 画面の見方", "## 画面構成",
		"go run ./cmd/trumpcards testgame\n# または\ngo run ./cmd/trumpcards --lang en testgame",
		"go run ./cmd/trumpcards web\n# または\ngo run ./cmd/server",
	).Replace(conformingManual)

	fake := &testing.T{}
	checkManual(fake, webSpec, "testgame", "テスト", base, "docs/manual/web/testgame.md")
	if fake.Failed() {
		t.Fatal("the Web control manual should conform; fix the fixture before trusting the mutations")
	}

	for _, c := range []struct {
		name string
		src  string
	}{
		{"cmd/server missing", strings.Replace(base, "go run ./cmd/server", "go run ./cmd/other", 1)},
		{"API route documented in prose", base + "\nPOST /testgame/exec で状態を進めます。\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			fake := &testing.T{}
			checkManual(fake, webSpec, "testgame", "テスト", c.src, "docs/manual/web/testgame.md")
			if !fake.Failed() {
				t.Errorf("the guard stayed silent on %q", c.name)
			}
		})
	}
}

// A permitted extra heading that merely starts with a required section's name
// must not read as a duplicate of it. The check used to count NUL-delimited
// substrings, so `## ルール一覧` beside `## ルール` would have failed a manual
// that is perfectly legal -- the templates explicitly allow game-specific
// sections.
func TestManualTemplateGuardAllowsHeadingsPrefixedByARequiredName(t *testing.T) {
	src := strings.Replace(conformingManual, "## ルール\n", "## ルール\n\n無し。\n\n## ルール一覧\n", 1)
	fake := &testing.T{}
	checkManual(fake, cuiSpec, "testgame", "テスト", src, "docs/manual/cui/testgame.md")
	if fake.Failed() {
		t.Error("the guard rejected a manual whose extra heading merely shares a prefix with a required one")
	}
}

// A `#` inside a fenced block is not a heading. Getting this wrong produced 13
// phantom "duplicate H1" reports when these manuals were first measured, so it
// gets its own test rather than riding on the cases above.
func TestManualHeadingsIgnoresFencedCode(t *testing.T) {
	h1, h2 := manualHeadings(conformingManual)
	if len(h1) != 1 {
		t.Errorf("got %d H1 headings %q, want 1 — a `#` comment inside a fence was counted", len(h1), h1)
	}
	if slices.Contains(h2, "これは見出しではない") {
		t.Errorf("a fenced `#` line was parsed as a heading: %v", h2)
	}
}

// manualBacktickRe matches one backtick-quoted span inside a table cell.
var manualBacktickRe = regexp.MustCompile("`([^`]+)`")

// manualGoStringRe matches a Go interpreted string literal, honouring escapes.
//
// The escape handling is load-bearing in both directions. `"[^"]*"` swallows an
// escaped quote; `"[^"\\]*"` desynchronises on the first literal containing a
// backslash and then pairs the wrong quotes for the rest of the file, which
// made a first pass report Cassino's `take` and `next` as undispatchable when
// both are declared a few lines apart.
var manualGoStringRe = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)

// manualBindCuiRe pairs a registered game with the controller that serves it.
var manualBindCuiRe = regexp.MustCompile(`BindCuiFor\("([a-z0-9]+)",(?s:.*?)controller\.New(\w+)CuiController`)

// manualUniversalCommands are dispatched before any controller sees them --
// execCuiCommand handles q/quit/exit and r/reset, handleCuiHintAndLog serves
// h/hint and log/l, and GameManager.Exec takes help/?/switch/games. None of
// them appears in a game's own source, so they are always documentable.
var manualUniversalCommands = map[string]bool{
	"r": true, "reset": true, "q": true, "quit": true, "exit": true,
	"help": true, "?": true, "h": true, "hint": true, "log": true, "l": true,
	"switch": true, "games": true,
}

// TestManualCommandsAreDispatchable asserts that every command a CUI manual
// documents is one the game actually accepts.
//
// TestPerGameManualsFollowTemplate checks the shape of the command table -- its
// three columns and the reset/quit/help rows -- and says nothing about whether
// the tokens inside it work. 19 manuals documented 28 tokens the dispatcher
// rejects (issue #5227): cego offered `p` where only `play` is accepted, kemps
// and spoons documented a `setdifficulty` command that exists nowhere in the
// game, and mighty's `pass` row described itself as equivalent to `bid 0` while
// not being dispatched at all. A reader copying those gets a "did you mean"
// suggestion, or in the several games that do not report unknown commands, a
// silently redrawn board.
//
// A command reaches the dispatcher only by appearing as a string literal in its
// controller, so literal membership is the test. That is deliberately loose --
// a literal used for something else would pass -- because the alternative
// (parsing each controller's switch) breaks on the games whose arguments are
// matched in a nested switch.
func TestManualCommandsAreDispatchable(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(repoRoot, "internal/infrastructure/ui/GameManager.go"))
	if err != nil {
		t.Fatalf("read GameManager.go: %v", err)
	}
	owner := map[string]string{}
	for _, m := range manualBindCuiRe.FindAllStringSubmatch(string(src), -1) {
		owner[m[1]] = m[2]
	}
	if len(owner) < len(games.All())-2 {
		t.Fatalf("found %d CUI bindings for %d games — BindCuiFor's shape changed; "+
			"fix manualBindCuiRe rather than trusting a clean run", len(owner), len(games.All()))
	}

	checked := 0
	for _, g := range games.All() {
		typ, ok := owner[g.Name]
		if !ok {
			continue
		}
		ctl, err := os.ReadFile(filepath.Join(repoRoot, "internal/adapter/controller", typ+"CuiController.go"))
		if err != nil {
			continue
		}
		literals := map[string]bool{}
		for _, q := range manualGoStringRe.FindAllString(string(ctl), -1) {
			if s, err := strconv.Unquote(q); err == nil {
				literals[s] = true
			}
		}
		rel := filepath.Join("docs/manual/cui", g.Name+".md")
		data, err := os.ReadFile(filepath.Join(repoRoot, rel))
		if err != nil {
			continue
		}
		body, ok := manualSection(string(data), "コマンド一覧")
		if !ok {
			continue
		}
		checked++
		for _, tok := range manualDocumentedCommands(body) {
			if manualUniversalCommands[tok] || literals[tok] {
				continue
			}
			t.Errorf("%s: documents `%s`, which %sCuiController never receives — "+
				"a command reaches the dispatcher only as a string literal there. "+
				"Correct the token or drop the row (issue #5227).", rel, tok, typ)
		}
	}
	if checked < len(games.All())-5 {
		t.Fatalf("only %d command tables were read for %d games — the walk broke", checked, len(games.All()))
	}
}

// manualDocumentedCommands returns the head token of the コマンド and 短縮形
// cells of every row in a command table.
func manualDocumentedCommands(body string) []string {
	var out []string
	seen := map[string]bool{}
	// Only the first contiguous run of table rows. Several sections follow the
	// command table with a second table describing its arguments -- speed's
	// `| idx | 出す手札のインデックス |`, sevens' `| cardIdx | ... |` -- and
	// reading those made the guard demand that `idx` and `cardIdx` be
	// dispatchable commands.
	started := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			if started {
				break
			}
			continue
		}
		started = true
		cells := splitManualRow(trimmed)
		for i, cell := range cells {
			if i > 1 {
				break
			}
			for _, span := range manualBacktickRe.FindAllStringSubmatch(cell, -1) {
				fields := strings.Fields(span[1])
				if len(fields) == 0 {
					continue
				}
				tok := fields[0]
				if tok == "コマンド" || tok == "短縮形" || seen[tok] {
					continue
				}
				seen[tok] = true
				out = append(out, tok)
			}
		}
	}
	return out
}

// splitManualRow splits a markdown table row, honouring `\|` escapes.
func splitManualRow(line string) []string {
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	var cells []string
	var cur strings.Builder
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' && i+1 < len(line) && line[i+1] == '|' {
			cur.WriteByte('|')
			i++
			continue
		}
		if line[i] == '|' {
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(line[i])
	}
	cells = append(cells, strings.TrimSpace(cur.String()))
	return cells
}

func TestManualDocumentedCommandsReadsOnlyTheCommandTable(t *testing.T) {
	// The argument-description table that follows several command tables must
	// not be read as more commands: speed documents `| idx | 出す手札の… |`
	// right below its command table, and treating `idx` as a command made the
	// guard demand that the controller dispatch it.
	body := "\n| コマンド | 短縮形 | 説明 |\n|---|---|---|\n| `play idx pile` | `p idx pile` | 出す |\n" +
		"\n`p idx pile` の各引数:\n\n| 引数 | 説明 |\n|---|---|\n| `idx` | 手札の位置 |\n| `pile` | 場札の位置 |\n"
	got := manualDocumentedCommands(body)
	if slices.Contains(got, "idx") || slices.Contains(got, "pile") {
		t.Errorf("read the argument table as commands: %v", got)
	}
	if !slices.Contains(got, "play") || !slices.Contains(got, "p") {
		t.Errorf("lost a real command: %v", got)
	}
}

func TestSplitManualRowHonoursEscapedPipes(t *testing.T) {
	// `bid entrar <s\|c\|h\|d>` is one cell, not four.
	got := splitManualRow("| `bid entrar <s\\|c\\|h\\|d>` | `b` | 宣言 |")
	if len(got) != 3 {
		t.Fatalf("got %d cells %q, want 3 — an escaped pipe was treated as a separator", len(got), got)
	}
	if got[0] != "`bid entrar <s|c|h|d>`" {
		t.Errorf("first cell is %q, want the unescaped command", got[0])
	}
}

func TestManualGoStringReSurvivesEscapes(t *testing.T) {
	// A literal containing a backslash must not desynchronise the ones after
	// it. `"[^"\\]*"` does, which made a first pass believe Cassino never
	// declares `take`.
	src := `fmt.Sprintf("line\n")` + "\n" + `[]string{"take", "t"}`
	var got []string
	for _, q := range manualGoStringRe.FindAllString(src, -1) {
		if s, err := strconv.Unquote(q); err == nil {
			got = append(got, s)
		}
	}
	if !slices.Contains(got, "take") || !slices.Contains(got, "t") {
		t.Errorf("lost a literal after an escaped one: %v", got)
	}
}
