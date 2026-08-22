//go:build test

package domain

// Seven Twenty-Seven のテスト専用ヘルパ。

// SetHandForTest は playerIdx の手札を丸ごと入れ替える（テスト用）。
func (g *SevenTwentySeven) SetHandForTest(playerIdx int, cards []*Card) {
	p := g.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// SetStandingForTest は playerIdx の「もう引かない」状態を設定する（テスト用）。
func (g *SevenTwentySeven) SetStandingForTest(playerIdx int, v bool) {
	g.players[playerIdx].SetStanding(v)
}

// SetPotForTest はポットを設定する（テスト用）。
func (g *SevenTwentySeven) SetPotForTest(pot int) { g.state.pot = pot }

// SettleForTest は決着処理を直接呼ぶ（テスト用）。
func (g *SevenTwentySeven) SettleForTest() { g.settle() }

// CpuDrawsForTest は CPU の判断を返す（テスト用）。
func (g *SevenTwentySeven) CpuDrawsForTest(playerIdx int) bool {
	return g.cpuDraws(g.players[playerIdx])
}

// StandEveryoneForTest は全員を「止まった」状態にする（テスト用）。
func (g *SevenTwentySeven) StandEveryoneForTest() {
	for _, p := range g.players {
		p.SetStanding(true)
	}
}
