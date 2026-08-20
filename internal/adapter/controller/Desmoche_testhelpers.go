//go:build test

package controller

// NewDesmocheDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewDesmocheDefaultOutputForTest(msg string) *DesmocheWebOutput {
	return newDesmocheDefaultOutput(msg)
}
