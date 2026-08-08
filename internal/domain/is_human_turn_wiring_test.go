//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// humanTurnGame is the slice of a game this test needs: whether it is the
// human's turn.
type humanTurnGame interface {
	IsHumanTurn() bool
}

// These 25 games' IsHumanTurn had no test at all. The bodies now delegate to
// the shared isHumanTurn helper (#5203), and codecov flagged the delegation as
// uncovered patch -- correctly, since the pre-existing bodies were untested
// too. The helper's own semantics are pinned in player_helpers_internal_test.go;
// what these assert is that each game is actually wired to it and answers
// without panicking on a freshly dealt game.
func TestIsHumanTurnWiredForEveryGame(t *testing.T) {
	t.Parallel()

	games := []struct {
		name string
		make func() humanTurnGame
	}{
		{"Belote", func() humanTurnGame { return NewDefaultBelote() }},
		{"Bezique", func() humanTurnGame { return NewDefaultBezique() }},
		{"BidWhist", func() humanTurnGame { return NewDefaultBidWhist() }},
		{"CourtPiece", func() humanTurnGame { return NewDefaultCourtPiece() }},
		{"Doppelkopf", func() humanTurnGame { return NewDefaultDoppelkopf() }},
		{"Ecarte", func() humanTurnGame { return NewDefaultEcarte() }},
		{"FiveHundred", func() humanTurnGame { return NewDefaultFiveHundred() }},
		{"FortyFives", func() humanTurnGame { return NewDefaultFortyFives() }},
		{"Kalooki", func() humanTurnGame { return NewDefaultKalooki() }},
		{"Klaverjas", func() humanTurnGame { return NewDefaultKlaverjas() }},
		{"KnockoutWhist", func() humanTurnGame { return NewDefaultKnockoutWhist() }},
		{"Manille", func() humanTurnGame { return NewDefaultManille() }},
		{"Mao", func() humanTurnGame { return NewDefaultMao() }},
		{"Marias", func() humanTurnGame { return NewDefaultMarias() }},
		{"Nap", func() humanTurnGame { return NewDefaultNap() }},
		{"Preference", func() humanTurnGame { return NewDefaultPreference() }},
		{"Sedma", func() humanTurnGame { return NewDefaultSedma() }},
		{"SoloWhist", func() humanTurnGame { return NewDefaultSoloWhist() }},
		{"SpoilFive", func() humanTurnGame { return NewDefaultSpoilFive() }},
		{"Sueca", func() humanTurnGame { return NewDefaultSueca() }},
		{"TeenPatti", func() humanTurnGame { return NewDefaultTeenPatti() }},
		{"ThreeCardBrag", func() humanTurnGame { return NewDefaultThreeCardBrag() }},
		{"Tressette", func() humanTurnGame { return NewDefaultTressette() }},
		{"Tute", func() humanTurnGame { return NewDefaultTute() }},
		{"TwentyNine", func() humanTurnGame { return NewDefaultTwentyNine() }},
	}

	for _, g := range games {
		t.Run(g.name, func(t *testing.T) {
			t.Parallel()
			game := g.make()
			assert.NotPanics(t, func() { _ = game.IsHumanTurn() },
				"a freshly constructed game must answer IsHumanTurn without panicking")
			// NewDefault* seats the human first and starts the turn there, so a
			// correctly wired delegation returns true here. Verified true for all
			// 25 before asserting it. NotPanics alone would also pass for a body
			// hardcoded to `return false`; this pins the answer as well as the
			// wiring. Raised in review on #5204.
			assert.True(t, game.IsHumanTurn(),
				"the human is seated first on a fresh game, so it is their turn")
		})
	}
}
