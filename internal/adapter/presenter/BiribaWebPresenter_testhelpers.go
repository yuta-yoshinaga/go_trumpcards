//go:build test

package presenter

// BiribaWebHintReasonKeyForTest exposes the reason mapping to tests in the
// external test package.
func BiribaWebHintReasonKeyForTest(reason string) string { return biribaWebHintReasonKeys[reason] }
