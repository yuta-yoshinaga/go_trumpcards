//go:build test

package domain

// SetContractForTest は契約と Solo の切り札スートを設定する (テスト用)。
func (g *Schafkopf) SetContractForTest(c SchafkopfContract, soloSuit int) {
	g.contract = c
	g.soloSuit = soloSuit
}

// SetPartnerRevealedForTest sets whether the partner has been revealed in tests.
func (g *Schafkopf) SetPartnerRevealedForTest(revealed bool) { g.partnerRevealed = revealed }
