//go:build test

package controller

// NewNainJauneDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewNainJauneDefaultOutputForTest(msg string) *NainJauneWebOutput {
	return newNainJauneDefaultOutput(msg)
}
