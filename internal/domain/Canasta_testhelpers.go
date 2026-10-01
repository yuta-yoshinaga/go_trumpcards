//go:build test

package domain

// CanastaScoreRoundForTest invokes the round scorer for domain tests.
func (g *Canasta) CanastaScoreRoundForTest(goOutPlayerIdx, goOutBonus int) {
	g.scoreRound(goOutPlayerIdx, goOutBonus)
}
