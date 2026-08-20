package ui

import "github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

// CuiHelpSpec describes the variable parts of a game's help text.
// The shared scaffold (gameCommands header, settings header, session header,
// reset/quit/help entries) is supplied by BuildCuiHelp.
type CuiHelpSpec struct {
	// Body, when non-empty, replaces the entire generated help text. Used by
	// games with hand-authored help that does not fit the standard scaffold
	// (and whose per-command lines are not yet i18n'd). Prefer the structured
	// fields below when possible; Body exists so every CUI entry can route
	// through cuiEntry(ctrl, spec) — one public API for the CLI wiring —
	// even when the help content is bespoke (issue #1460).
	Body []string
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
	// ExampleKeys are i18n keys for a worked sequence of commands, in the order
	// a player would type them. Optional: when empty the examples section
	// (header + entries) is omitted entirely, so a game that supplies none
	// renders exactly as it did before the section existed.
	//
	// The command tables above say what each command means; they do not say
	// which one to type first, and with 318 games each carrying its own
	// vocabulary that is the question a newcomer actually has (issue #5358).
	ExampleKeys []string
	// ResetOverride, when non-empty, replaces the default i18n.T("resetEntry")
	// line in the session section. Used by games whose reset command accepts
	// options (e.g. Sevens: "r [tunnel] [joker=N] [strategy] [passes=N]").
	// Extracted per issue #1460 so such games can use the standard template
	// instead of bypassing BuildCuiHelp with a raw []string.
	ResetOverride string
	// NoteKeys are i18n keys for rules the player needs while reading the
	// command list but that are not commands themselves -- e.g. Macau's magic
	// cards (2/7/8/J), which the Web GUI shows as a permanent reference table
	// while the CUI said nothing at all (#5622).
	//
	// Rendered last, after the examples: the tables answer "what can I type",
	// the examples answer "what do I type first", and these answer "what just
	// happened to me".
	NoteKeys []string
}

// BuildCuiHelp assembles standard CUI help text from spec. Order:
// title, blank, gameCommands header, command keys, extra command lines,
// blank + settings header + setting keys (only when non-empty),
// blank + session header + reset + quit + help,
// blank + examples header + example keys (only when non-empty).
// When spec.Body is non-empty the scaffold is skipped entirely and Body
// is returned verbatim.
//
// Examples come last, after the session block: the command tables are the
// reference a returning player scans, while the worked sequence is for someone
// who does not yet know which command to type first.
func BuildCuiHelp(spec CuiHelpSpec) []string {
	if len(spec.Body) > 0 {
		return spec.Body
	}
	lines := make([]string, 0, 10+len(spec.CommandKeys)+len(spec.ExtraCommandLines)+len(spec.SettingKeys)+len(spec.ExtraSettingLines)+len(spec.NoteKeys))
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
	if len(spec.ExampleKeys) > 0 {
		lines = append(lines, "", i18n.T("examples"))
		for _, k := range spec.ExampleKeys {
			lines = append(lines, i18n.T(k))
		}
	}
	if len(spec.NoteKeys) > 0 {
		lines = append(lines, "", i18n.T("notes"))
		for _, k := range spec.NoteKeys {
			lines = append(lines, i18n.T(k))
		}
	}
	return lines
}
