package controller

import "github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

// invalidArg renders a rejected-argument message with the error marker that
// unknownCommandMessage already applies.
//
// The marker is not decoration. Without it a rejection is byte-for-byte an
// ordinary reply, so nothing downstream -- the CUI's own error styling, or a
// test executing a documented example -- can tell "the game refused this" from
// "the game did this".
//
// Measured on 2026-08-15 by giving a malformed argument to every one of the 638
// argument-taking commands in the help tables, on a freshly dealt game: 4 came
// back marked. The rest returned a perfectly good message that read as success,
// or (27 commands) nothing but a redrawn board. See issue #5377.
//
// The signature deliberately mirrors i18n.Tf so the call sites differ only in
// the function name.
func invalidArg(key string, params ...string) string {
	return i18n.MarkError(i18n.Tf(key, params...))
}
