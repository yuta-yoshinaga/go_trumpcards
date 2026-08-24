//go:build test

package domain

// This file contains test helper methods for the mixed-game orchestrator.
// They exist solely for cross-package test setup and are not part of the
// production game logic.

// SetDisciplineForTest jumps the table to a discipline and deals a hand of it
// (test only).
//
// **The rotation is what production changes.** Walking a match to the eighth
// discipline takes dozens of hands and depends on the deal, so tests that need
// one specific discipline start it directly instead.
func (g *Horse) SetDisciplineForTest(d HorseDiscipline) {
	g.discipline = d
	g.startHand()
}
