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
// Measured by giving a malformed argument to every argument-taking command in
// the help tables, on a freshly dealt game. When this helper was introduced on
// 2026-08-15, 4 of 638 came back marked; after the i18n migration moved most
// rejections onto cuiutil.ParseIntArgKeys and that path started marking too,
// it is 668 of 994. The rest still return a message that reads as success, or
// (45 commands) nothing but a redrawn board -- issues #5377 and #5390.
//
// The signature deliberately mirrors i18n.Tf so the call sites differ only in
// the function name.
func invalidArg(key string, params ...string) string {
	return i18n.MarkError(i18n.Tf(key, params...))
}
