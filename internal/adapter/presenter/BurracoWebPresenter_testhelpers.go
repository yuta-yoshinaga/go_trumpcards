//go:build test

package presenter

// BurracoWebHintReasonKeyForTest exposes the reason mapping to tests in the
// external test package.
func BurracoWebHintReasonKeyForTest(reason string) string { return burracoWebHintReasonKeys[reason] }
