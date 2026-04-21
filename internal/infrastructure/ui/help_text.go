package ui

import "github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

// CuiHelpSpec describes the variable parts of a game's help text.
// The shared scaffold (gameCommands header, settings header, session header,
// reset/quit/help entries) is supplied by BuildCuiHelp.
type CuiHelpSpec struct {
	// TitleKey is the i18n key for the help title line (e.g. "hearts.helpTitle").
	TitleKey string
	// CommandKeys are i18n keys for game-specific command lines, in display order.
	CommandKeys []string
	// ExtraCommandLines are literal lines appended after CommandKeys but before
	// the settings section. Used for entries not yet in i18n (e.g. action log,
	// clear-history) whose spacing is per-game column-aligned.
	ExtraCommandLines []string
	// SettingKeys are i18n keys for game-specific setting lines. If both this
	// and ExtraSettingLines are empty the settings section (header + entries)
	// is omitted entirely.
	SettingKeys []string
	// ExtraSettingLines are literal lines appended after SettingKeys in the
	// settings section. Used for settings not yet in i18n.
	ExtraSettingLines []string
	// ResetOverride, when non-empty, replaces the default i18n.T("resetEntry")
	// line in the session section. Used by games whose reset command accepts
	// options (e.g. Sevens: "r [tunnel] [joker=N] [strategy] [passes=N]").
	// Extracted per issue #1460 so such games can use the standard template
	// instead of bypassing BuildCuiHelp with a raw []string.
	ResetOverride string
}

// BuildCuiHelp assembles standard CUI help text from spec. Order:
// title, blank, gameCommands header, command keys, extra command lines,
// blank + settings header + setting keys (only when non-empty),
// blank + session header + reset + quit + help.
func BuildCuiHelp(spec CuiHelpSpec) []string {
	// 10 fixed lines: title + blank + gameCommands header + blank + settings header +
	// blank + session header + reset + quit + help. The two "blank + settings header" lines
	// are included unconditionally in capacity; the slight over-allocation when the settings
	// section is skipped is harmless.
	lines := make([]string, 0, 10+len(spec.CommandKeys)+len(spec.ExtraCommandLines)+len(spec.SettingKeys)+len(spec.ExtraSettingLines))
	lines = append(lines, i18n.T(spec.TitleKey), "", i18n.T("gameCommands"))
	for _, k := range spec.CommandKeys {
		lines = append(lines, i18n.T(k))
	}
	lines = append(lines, spec.ExtraCommandLines...)
	if len(spec.SettingKeys) > 0 || len(spec.ExtraSettingLines) > 0 {
		lines = append(lines, "", i18n.T("settings"))
		for _, k := range spec.SettingKeys {
			lines = append(lines, i18n.T(k))
		}
		lines = append(lines, spec.ExtraSettingLines...)
	}
	resetLine := i18n.T("resetEntry")
	if spec.ResetOverride != "" {
		resetLine = spec.ResetOverride
	}
	lines = append(lines,
		"",
		i18n.T("session"),
		resetLine,
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	)
	return lines
}
