//go:build test

package domain

// This file contains test helper methods for UltimateTexasHoldem.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetPayouts sets payout fields for deterministic round accounting tests.
func (u *UltimateTexasHoldem) SetPayouts(ante, blind, play, trips int) {
	u.antePayout, u.blindPayout, u.playPayout, u.tripsPayout = ante, blind, play, trips
}
