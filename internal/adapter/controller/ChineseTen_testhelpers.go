//go:build test

package controller

// NewChineseTenDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
//
// Lives behind the `test` tag so it never reaches a shipped binary — the same
// convention as internal/domain/*_testhelpers.go. The category half of the
// constraint is copied from ChineseTenWebController.go: ChineseTenWebOutput is
// declared there, so a bare `test` tag would break any build that has `test`
// set without this game's category.
func NewChineseTenDefaultOutputForTest(msg string) *ChineseTenWebOutput {
	return newChineseTenDefaultOutput(msg)
}
