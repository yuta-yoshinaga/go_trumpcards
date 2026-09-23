//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Jass) CardRankPublic(card *Card) int { return g.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Jass) CardPointsPublic(card *Card) int { return jassCardPoints(card, g.trumpSuit) }

// ResolveWeisForTest sets the trump/forehand and runs Weis resolution (テスト用)。
func (g *Jass) ResolveWeisForTest(trumpSuit, forehandIdx int) {
	g.trumpSuit = trumpSuit
	g.forehandIdx = forehandIdx
	g.weisResolved = false
	g.resolveWeis()
}

// ResolveStockForTest sets the trump and runs Stöck resolution (テスト用)。
func (g *Jass) ResolveStockForTest(trumpSuit int) {
	g.trumpSuit = trumpSuit
	g.resolveStock()
}

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Jass) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < JassTeamCnt {
		g.roundPoints[team] += pts
	}
}

// SetBidPlayerIdxForTest sets the active bid player (テスト用)。
func (g *Jass) SetBidPlayerIdxForTest(idx int) { g.bidPlayerIdx = idx }

// SetDealerIdxForTest sets the dealer (テスト用)。
func (g *Jass) SetDealerIdxForTest(idx int) { g.dealerIdx = idx }

// RebeginRoundForTest re-deals and re-enters the bid phase (テスト用)。
func (g *Jass) RebeginRoundForTest() {
	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}
