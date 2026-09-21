//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuitKeyOf(t *testing.T) {
	for _, tc := range []struct {
		suit int
		want string
	}{
		{CardDesignSpade, "common.suit.spade"},
		{CardDesignClover, "common.suit.club"},
		{CardDesignHeart, "common.suit.heart"},
		{CardDesignDiamond, "common.suit.diamond"},
		{-1, "common.suit.unknown"},
	} {
		assert.Equal(t, tc.want, suitKeyOf(tc.suit))
	}
}
