//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrescentFoundationHelpers(t *testing.T) {
	wantSuits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for i := range CrescentFoundationCnt {
		assert.Equal(t, wantSuits[i%CrescentAscendingFoundationCnt], crescentFoundationSuit(i))
	}
}
