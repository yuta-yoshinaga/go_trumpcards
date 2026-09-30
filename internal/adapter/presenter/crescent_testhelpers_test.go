//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

func testCrescentFoundationSuit(i int) int {
	return []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}[i%4]
}
