package controller

import "github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

// invalidArg renders a rejected-argument message with the error marker that
// unknownCommandMessage already applies.
//
// The marker is not decoration. Without it a rejection is byte-for-byte an
// ordinary reply, so nothing downstream -- the CUI's own error styling, or a
// test executing a documented example -- can tell "the game refused this" from
// "the game did this". Measured on 2026-08-15, 2 of the 291 games with an
// argument-taking command marked such a rejection; the other 289 returned a
// perfectly good message (or, for 27 of them, nothing but a redrawn board)
// that read as success. See issue #5377.
//
// The signature deliberately mirrors i18n.Tf so the call sites differ only in
// the function name.
func invalidArg(key string, params ...string) string {
	return i18n.MarkError(i18n.Tf(key, params...))
}
