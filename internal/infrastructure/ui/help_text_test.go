//go:build test

package ui

import (
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
