//go:build test

package domain

// This file contains test helper functions for Tarocco Piemontese. They exist
// solely for cross-package test setup and are not part of the game logic.

// NewPiedmonteseTarotPlayersForTest builds a seat list of the given size
// (seat 0 human), for tests that construct a table directly.
func NewPiedmonteseTarotPlayersForTest(seats int) []*PiedmonteseTarotPlayer {
	return newPiedmonteseTarotPlayers(seats)
}

// PiedmonteseTarotCpuScartoForTest returns the cards the CPU logic would bury
// for the given seat, so a test can drive the human dealer through the scarto
// without re-implementing the rule.
func PiedmonteseTarotCpuScartoForTest(g *PiedmonteseTarot, playerIdx int) []int {
	return g.cpuSelectScarto(playerIdx)
}

// SetDealerForTest moves the deal to another seat and re-deals (test only).
func (g *PiedmonteseTarot) SetDealerForTest(idx int) {
	g.dealerIdx = idx
	g.startRound()
}
