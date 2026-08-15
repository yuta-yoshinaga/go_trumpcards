//go:build test

package controller

// NewLaughAndLieDownDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewLaughAndLieDownDefaultOutputForTest(msg string) *LaughAndLieDownWebOutput {
	return newLaughAndLieDownDefaultOutput(msg)
}
