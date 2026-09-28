//go:build test

package domain

// IsTrumpPublic テスト用公開ラッパー。
func (g *Watten) IsTrumpPublic(c *Card) bool { return g.isTrump(c) }

// CardRankPublic テスト用公開メソッド。
func (g *Watten) CardRankPublic(c *Card) int { return g.cardRank(c) }

// SetupRaiseForTest configures a pending-raise/respond state (テスト用)。
func (g *Watten) SetupRaiseForTest(pending, raiserTeam, responderIdx int) {
	g.phase = WattenPhaseRespond
	g.pendingStake = pending
	g.raiserTeam = raiserTeam
	g.responderIdx = responderIdx
}

// SetTeamTricksForTest sets a team's trick count for the current deal (テスト用)。
func (g *Watten) SetTeamTricksForTest(team, n int) {
	if team >= 0 && team < WattenTeamCnt {
		g.teamTricks[team] = n
	}
}

// SetRaiseCountForTest sets the accepted raise count (テスト用)。
func (g *Watten) SetRaiseCountForTest(n int) { g.raiseCount = n }
