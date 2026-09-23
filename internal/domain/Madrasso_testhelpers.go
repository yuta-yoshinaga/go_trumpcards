//go:build test

package domain

// SetTeamRoundPointsForTest はチームの現ラウンド獲得点を設定する (テスト用)。
func (g *Madrasso) SetTeamRoundPointsForTest(team, pts int) {
	if team >= 0 && team < MadrassoTeamCnt {
		g.teamRoundPoints[team] = pts
	}
}

// SetTrumpSuitForTest は切り札スートを設定する (テスト用)。
func (g *Madrasso) SetTrumpSuitForTest(suit int) { g.trumpSuit = suit }

// MadrassoStrengthForTest は札位の強さを返す (テスト用)。
func MadrassoStrengthForTest(value int) int { return madrassoStrength(value) }

// MadrassoPointsForTest はカード点を返す (テスト用)。
func MadrassoPointsForTest(value int) int { return madrassoPoints(value) }
