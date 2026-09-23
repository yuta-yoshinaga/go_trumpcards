//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Bauernschnapsen) CardRankPublic(card *Card) int { return BauernschnapsenRankOrder(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Bauernschnapsen) CardPointsPublic(card *Card) int { return BauernschnapsenCardPoints(card) }

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Bauernschnapsen) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < BauernschnapsenTeamCnt {
		g.roundPoints[team] += pts
	}
}

// SetContractForTest はテスト用に契約と宣言者を設定する。
func (g *Bauernschnapsen) SetContractForTest(c BauernschnapsenContract, declarerIdx int) {
	g.contract = c
	g.declarerIdx = declarerIdx
}

// SetSeatTricksForTest はテスト用に席別の獲得トリック数を設定する。
func (g *Bauernschnapsen) SetSeatTricksForTest(idx, tricks int) {
	if idx < 0 || idx >= BauernschnapsenPlayerCnt {
		return
	}
	g.seatTricks[idx] = tricks
}

// SetRoundResultForTest はテスト用にチームのラウンド成績を設定する。
func (g *Bauernschnapsen) SetRoundResultForTest(team, points, tricks int) {
	if team < 0 || team >= BauernschnapsenTeamCnt {
		return
	}
	g.roundPoints[team] = points
	g.roundTricks[team] = tricks
}

// ContractMadeForTest はテスト用に契約の成否を返す。
func (g *Bauernschnapsen) ContractMadeForTest(declarerTeam int) bool {
	return g.contractMade(declarerTeam)
}
