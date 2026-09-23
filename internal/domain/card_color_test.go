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

func TestTrumpKeyOf(t *testing.T) {
	for _, tc := range []struct {
		suit int
		want string
	}{
		{CardDesignSpade, "common.suit.spade"},
		{CardDesignClover, "common.suit.club"},
		{CardDesignHeart, "common.suit.heart"},
		{CardDesignDiamond, "common.suit.diamond"},
		{0, "common.suit.notrump"},
		{99, "common.suit.notrump"},
	} {
		assert.Equal(t, tc.want, trumpKeyOf(tc.suit))
	}
}
