//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBuraWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so an ordinary reset arrives with a nil
	// *BuraWebConfig. Calling the method on that nil pointer panicked and
	// returned a 500 for the very request that starts a game -- caught by E2E,
	// not by any unit test, which is why this one exists.
	var input BuraWebInput
	assert.NotPanics(t, func() {
		cfg := input.ToConfig()
		assert.Equal(t, domain.DefaultBuraConfig(), cfg)
	})
}

func TestBuraWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	input := BuraWebInput{Config: &BuraWebConfig{CpuDifficulty: &bad}}
	cfg := input.ToConfig()
	assert.NoError(t, cfg.Validate(), "an out-of-range difficulty must be clamped, not passed through")
}

func TestNewBuraDefaultOutput_CarriesTheWinThreshold(t *testing.T) {
	// The page reads winThreshold off the response rather than hardcoding 31;
	// an error response that omits it would render "claim 0".
	out := newBuraDefaultOutput("boom")
	assert.Equal(t, domain.BuraWinThreshold, out.WinThreshold)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.CurrentLead)
}
