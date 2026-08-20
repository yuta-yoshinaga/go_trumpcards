//go:build test

package controller

// NewSkitgubbeDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go.
func NewSkitgubbeDefaultOutputForTest(msg string) *SkitgubbeWebOutput {
	return newSkitgubbeDefaultOutput(msg)
}
