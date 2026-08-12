//go:build test

package controller

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SnapConfigFromInputForTest exposes snapConfigFromInput to the external test package.
func SnapConfigFromInputForTest(cur domain.SnapConfig, in *SnapWebConfig) domain.SnapConfig {
	return snapConfigFromInput(cur, in)
}
