//go:build test

package ui

import (
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
