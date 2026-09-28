//go:build test

package domain

// ResolveRowsForTest は所定の列を注入してラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決・Refait ロジックを決定的に検証するためのショートカット。
func (g *TrenteEtQuarante) ResolveRowsForTest(bet TrenteEtQuaranteBet, stake int, noir, rouge []*Card) {
	g.state.currentBet = bet
	g.state.stake = stake
	g.state.noirRow = noir
	g.state.rougeRow = rouge
	g.state.noirTotal = 0
	for _, c := range noir {
		g.state.noirTotal += trenteEtQuaranteCardValue(c)
	}
	g.state.rougeTotal = 0
	for _, c := range rouge {
		g.state.rougeTotal += trenteEtQuaranteCardValue(c)
	}
	if len(noir) > 0 {
		g.state.firstCardRed = trenteEtQuaranteIsRed(noir[0])
	}
	g.resolve()
}
