//go:build test

package controller

// NewToepenDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewToepenDefaultOutputForTest(msg string) *ToepenWebOutput { return newToepenDefaultOutput(msg) }
