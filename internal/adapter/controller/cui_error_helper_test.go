//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TestInvalidArgMarksTheMessage pins what invalidArg adds, because the tests of
// the individual controllers now spell their expectations with invalidArg too.
// That keeps them tied to production, but on its own it would also pass if
// invalidArg stopped marking anything -- so the marking is asserted here
// against i18n directly, not against invalidArg.
func TestInvalidArgMarksTheMessage(t *testing.T) {
	got := invalidArg("invalidIndex", "val", "zz")

	body, isErr := i18n.StripErrorPrefix(got)
	assert.True(t, isErr, "invalidArg must mark its message as an error")
	assert.Equal(t, i18n.Tf("invalidIndex", "val", "zz"), body,
		"stripping the marker must give back exactly the i18n text")
	assert.NotEqual(t, i18n.Tf("invalidIndex", "val", "zz"), got,
		"an unmarked message is indistinguishable from an accepted command")
}

// TestInvalidArgForwardsParams guards the variadic hand-off: a signature change
// that dropped the params would still produce a marked string, and every
// controller test comparing invalidArg against invalidArg would still pass.
func TestInvalidArgForwardsParams(t *testing.T) {
	body, _ := i18n.StripErrorPrefix(invalidArg("invalidIndex", "val", "sentinel-42"))
	assert.Contains(t, body, "sentinel-42")
}
