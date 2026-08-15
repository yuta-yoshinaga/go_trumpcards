//go:build test

package controller

// NewChineseTenDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewChineseTenDefaultOutputForTest(msg string) *ChineseTenWebOutput {
	return newChineseTenDefaultOutput(msg)
}
