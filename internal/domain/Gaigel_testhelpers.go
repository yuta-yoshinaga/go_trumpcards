//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Gaigel) CardRankPublic(card *Card) int { return GaigelRankOrder(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Gaigel) CardPointsPublic(card *Card) int { return GaigelCardPoints(card) }

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Gaigel) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < GaigelTeamCnt {
		g.roundPoints[team] += pts
	}
}

// RebeginRoundForTest re-deals and re-enters the play phase (テスト用)。
func (g *Gaigel) RebeginRoundForTest() {
	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}
