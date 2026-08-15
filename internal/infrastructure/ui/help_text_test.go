//go:build test

package ui

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBuildCuiHelp_includesScaffolding(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "hearts.helpTitle",
		CommandKeys: []string{"hearts.helpPass"},
		SettingKeys: []string{"hearts.helpSetLimit"},
	})

	want := []string{
		i18n.T("hearts.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("hearts.helpPass"),
		"",
		i18n.T("settings"),
		i18n.T("hearts.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
	assertLines(t, got, want)
}

func TestBuildCuiHelp_omitsSettingsWhenEmpty(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "baccarat.helpTitle",
		CommandKeys: []string{"baccarat.helpBet"},
		ExtraCommandLines: []string{
			"  log                  action log",
		},
	})

	want := []string{
		i18n.T("baccarat.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("baccarat.helpBet"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
	assertLines(t, got, want)
}

func TestBuildCuiHelp_extraLinesAppearBeforeSettings(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "spades.helpTitle",
		CommandKeys:       []string{"spades.helpBid", "spades.helpPlay"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"spades.helpSetDifficulty"},
	})

	// Find "settings" header in output and verify action log appears before it.
	settingsIdx := -1
	actionLogIdx := -1
	for i, line := range got {
		if line == i18n.T("settings") {
			settingsIdx = i
		}
		if line == "  l                    action log" {
			actionLogIdx = i
		}
	}
	if settingsIdx == -1 || actionLogIdx == -1 {
		t.Fatalf("missing markers: settingsIdx=%d actionLogIdx=%d (got=%v)", settingsIdx, actionLogIdx, got)
	}
	if actionLogIdx >= settingsIdx {
		t.Errorf("action log (%d) should precede settings header (%d)", actionLogIdx, settingsIdx)
	}
}

func TestBuildCuiHelp_settingsIncludesExtraLines(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "holdem.helpTitle",
		CommandKeys:       []string{"holdem.helpFold"},
		SettingKeys:       []string{"holdem.helpBettingLimit"},
		ExtraSettingLines: []string{"  sb <amount>          small blind (>=1)"},
	})

	want := []string{
		i18n.T("holdem.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("holdem.helpFold"),
		"",
		i18n.T("settings"),
		i18n.T("holdem.helpBettingLimit"),
		"  sb <amount>          small blind (>=1)",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
	assertLines(t, got, want)
}

func TestBuildCuiHelp_settingsOnlyExtraLines(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "test.helpTitle",
		CommandKeys:       []string{"test.helpCmd"},
		ExtraSettingLines: []string{"  x <n>    custom setting"},
	})

	// Expect "", settings header, extra setting line, "", session, ...
	want := []string{
		i18n.T("test.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("test.helpCmd"),
		"",
		i18n.T("settings"),
		"  x <n>    custom setting",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
	assertLines(t, got, want)
}

func TestBuildCuiHelp_bodyShortCircuitsScaffold(t *testing.T) {
	// Body wins over every other field — used by games whose help is
	// hand-authored and does not fit the standard scaffold (issue #1460).
	body := []string{
		"Canasta (カナスタ) Help",
		"",
		"Game Commands:",
		"  ds                   draw from stock",
	}
	got := BuildCuiHelp(CuiHelpSpec{
		Body:        body,
		TitleKey:    "hearts.helpTitle", // ignored
		CommandKeys: []string{"hearts.helpPass"},
		SettingKeys: []string{"hearts.helpSetLimit"},
	})
	assertLines(t, got, body)
}

// TestPokerHelpUsesI18n covers issue #1511: tournament rebuy/blind and stud
// ante/bring-in help lines must come from i18n keys so JA users see localized
// text instead of the English literals that used to live in ExtraCommandLines /
// ExtraSettingLines. Locale is restored to "en" (matching TestMain) at the end.
func TestPokerHelpUsesI18n(t *testing.T) {
	defer i18n.SetLang("en")

	tests := []struct {
		name        string
		game        string // registry name
		lang        string
		mustContain []string
	}{
		{
			name: "holdem JA shows tournament + blind translations",
			game: "holdem",
			lang: "ja",
			mustContain: []string{
				"リバイ",             // tournament.helpRebuy
				"アドオン",            // tournament.helpAddOn
				"スモールブラインド (>=1)", // tournament.helpSmallBlind
				"ブラインドレベルアップ手数 (>=1)", // tournament.helpLevelUpHands
				"テーブルサイズ",             // tournament.helpTableSize
			},
		},
		{
			name: "holdem EN keeps existing English wording",
			game: "holdem",
			lang: "en",
			mustContain: []string{
				"rebuy",
				"add-on",
				"small blind (>=1)",
				"blind level-up hands (>=1)",
				"table size",
			},
		},
		{
			name: "pineapple JA includes localized discard line",
			game: "pineapple",
			lang: "ja",
			mustContain: []string{
				"手札を捨てる", // tournament.helpDiscard
				"リバイ",
			},
		},
		{
			// Pinned separately from pineapple so a copy-paste swap of
			// tournamentRebuyAddOnKeys for pineappleRebuyAddOnKeys here would
			// fail this case loudly instead of going unnoticed.
			name: "crazypineapple JA includes the discard line too",
			game: "crazypineapple",
			lang: "ja",
			mustContain: []string{
				"手札を捨てる",
				"リバイ",
			},
		},
		{
			name: "sevencardstud JA shows ante and bring-in translations",
			game: "sevencardstud",
			lang: "ja",
			mustContain: []string{
				"アンテ (>=1)",         // stud.helpAnte
				"ブリングイン (>=1)",      // stud.helpBringIn
				"スモールベット (>=1)",     // stud.helpSmallBet
				"ビッグベット (>=1)",      // stud.helpBigBet
				"アンテレベルアップ手数 (>=1)", // stud.helpLevelUpHands
			},
		},
		{
			name: "razz JA shares the stud ante block",
			game: "razz",
			lang: "ja",
			mustContain: []string{
				"アンテ (>=1)",
				"ブリングイン (>=1)",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i18n.SetLang(tc.lang)
			var entry *GameRegistryEntry
			for i := range gameRegistry {
				if gameRegistry[i].Name == tc.game {
					entry = &gameRegistry[i]
					break
				}
			}
			if entry == nil {
				t.Fatalf("game %q not found in registry", tc.game)
			}
			joined := strings.Join(entry.NewCui().HelpLines(), "\n")
			for _, want := range tc.mustContain {
				if !strings.Contains(joined, want) {
					t.Errorf("help for %s (lang=%s) missing %q\nfull help:\n%s", tc.game, tc.lang, want, joined)
				}
			}
		})
	}
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d want %d\ngot=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q want %q", i, got[i], want[i])
		}
	}
}

// The examples section is opt-in: a game that supplies no example keys renders
// exactly as before. That is what lets the mechanism ship for all 318 games
// while only a few carry content — see issue #5358.
func TestBuildCuiHelp_omitsExamplesWhenEmpty(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "hearts.helpTitle",
		CommandKeys: []string{"hearts.helpPass"},
	})

	for _, line := range got {
		if line == i18n.T("examples") {
			t.Fatalf("examples header rendered for a game with no example keys:\n%v", got)
		}
	}
}

// Examples go last, after the session block: the command tables are the
// reference a returning player scans, and the worked sequence is for someone
// who does not yet know which command to type first.
func TestBuildCuiHelp_examplesComeAfterSession(t *testing.T) {
	got := BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "blackjack.helpTitle",
		CommandKeys: []string{"blackjack.helpHit"},
		ExampleKeys: []string{"blackjack.helpExampleBet", "blackjack.helpExampleHit"},
	})

	want := []string{
		i18n.T("blackjack.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("blackjack.helpHit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
		"",
		i18n.T("examples"),
		i18n.T("blackjack.helpExampleBet"),
		i18n.T("blackjack.helpExampleHit"),
	}
	assertLines(t, got, want)
}

// Body short-circuits the whole scaffold, examples included: those games hand
// author their help and must not gain a section they did not ask for.
func TestBuildCuiHelp_bodyStillShortCircuitsExamples(t *testing.T) {
	body := []string{"custom line"}
	got := BuildCuiHelp(CuiHelpSpec{Body: body, ExampleKeys: []string{"blackjack.helpExampleBet"}})
	assertLines(t, got, body)
}

// TestCuiHelpExamplesUseRealCommands asserts that every command shown in a
// game's examples section is one that game's own command table lists.
//
// Examples are prose in a locale file: nothing stops one naming a command the
// game does not have, and the result renders perfectly. Writing the first six
// by hand produced four such mistakes — hearts said `p 0 3 7` where the pass
// command is `pass`, gofish said `a 1 7` where it is `ask`, klondike invented
// `f 1` for what is `m w f`, and daifugo used a bare index where the play
// command is `p`. All four looked right until this check ran. See issue #5358.
//
// Reads the rendered help rather than the spec: the spec is turned into lines
// by BuildCuiHelp at registration time and not retained, and the rendered text
// is what a player actually sees.
func TestCuiHelpExamplesUseRealCommands(t *testing.T) {
	registry := GameRegistry()
	if len(registry) < 300 {
		t.Fatalf("only %d games in the registry -- the walk broke", len(registry))
	}

	// Both languages: the example and the command table are separate strings in
	// separate locale files, so a mismatch can exist in one and not the other.
	// Checking only the default language is how the first version of this guard
	// passed while ja/gofish.json still said `a 1 7` -- the tests run in en.
	original := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(original) })

	withExamples, checked := 0, 0
	var bad []string
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, entry := range registry {
			// HelpLines() is evaluated per NewCui() call, so this picks up the
			// language set above rather than the one at registration time.
			commands, examples := helpSections(entry.NewCui().HelpLines())
			if len(examples) == 0 {
				continue
			}
			withExamples++
			verbs := map[string]bool{}
			for _, line := range commands {
				if v := helpLineVerb(line); v != "" {
					verbs[v] = true
				}
			}
			for _, line := range examples {
				v := helpLineVerb(line)
				if v == "" {
					continue
				}
				checked++
				if !verbs[v] {
					bad = append(bad, lang+"/"+entry.Name+": example uses `"+v+"`, which its command table does not list")
				}
			}
		}
	}

	if withExamples == 0 {
		t.Fatal("no game rendered an examples section in either language -- either none wires ExampleKeys or helpSections stopped matching")
	}
	if checked == 0 {
		t.Fatalf("%d games have examples but no command token was parsed from them -- helpLineVerb stopped matching", withExamples)
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("examples naming commands the game does not have (%d bad of %d checked across %d games):\n  %s",
			len(bad), checked, withExamples, strings.Join(bad, "\n  "))
	}
}

// helpSections splits rendered help into its command lines and its example
// lines, keyed off the localized section headers so it follows a language
// switch rather than hard-coding Japanese or English.
func helpSections(lines []string) (commands, examples []string) {
	cmdHeader, exHeader := i18n.T("gameCommands"), i18n.T("examples")
	section := ""
	for _, line := range lines {
		switch line {
		case cmdHeader:
			section = "cmd"
			continue
		case exHeader:
			section = "ex"
			continue
		}
		if strings.TrimSpace(line) == "" {
			section = ""
			continue
		}
		switch section {
		case "cmd":
			commands = append(commands, line)
		case "ex":
			examples = append(examples, line)
		}
	}
	return commands, examples
}

// helpLineVerb extracts the command token from a help line such as
// "  b <n> ...  bet". Returns "" for a line that does not start with one (a
// bare index, a continuation line, a localized prose fragment).
func helpLineVerb(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	verb := fields[0]
	for _, r := range verb {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return ""
		}
	}
	// A bare number is an index, not a command.
	if _, err := strconv.Atoi(verb); err == nil {
		return ""
	}
	return verb
}
