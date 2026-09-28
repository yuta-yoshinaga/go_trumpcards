//go:build test

package domain

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (s *ShengJi) SetPhaseForTest(p ShengJiPhase) { s.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (s *ShengJi) SetHandForTest(idx int, cards []*Card) {
	p := s.GetPlayer(idx)
	if p == nil {
		return
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// SetLevelForTest はこの局のレベルを差し替える (テスト専用)。
func (s *ShengJi) SetLevelForTest(level int) { s.level = level }

// SetTeamLevelForTest はチームのレベルを差し替える (テスト専用)。
func (s *ShengJi) SetTeamLevelForTest(team, level int) {
	if team < 0 || team >= ShengJiTeamCnt {
		return
	}
	s.levels[team] = level
}

// SetTrumpForTest は切札スートを差し替える (テスト専用)。
func (s *ShengJi) SetTrumpForTest(suit int) { s.trumpSuit = suit }

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (s *ShengJi) SetCurrentPlayerForTest(idx int) {
	s.currentIdx = idx
	s.trickLeader = idx
}

// SetKittyForTest は底牌を差し替える (テスト専用)。
func (s *ShengJi) SetKittyForTest(cards []*Card) { s.kitty = cards }

// FinishHandForTest は局の精算を走らせる (テスト専用)。
func (s *ShengJi) FinishHandForTest() { s.finishHand() }
