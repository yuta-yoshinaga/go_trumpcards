//go:build test

package domain

// Mendikot のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (m *Mendikot) SetPhaseForTest(phase MendikotPhase) { m.phase = phase }

// SetTrickNumberForTest トリック番号を設定する
func (m *Mendikot) SetTrickNumberForTest(n int) { m.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (m *Mendikot) SetCurrentPlayerIdxForTest(i int) { m.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest リードプレイヤーを設定する
func (m *Mendikot) SetLeadPlayerIdxForTest(i int) { m.leadPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (m *Mendikot) SetCurrentTrickForTest(tc []*TrickCard) { m.currentTrick = tc }

// SetTrumpForTest 切り札と決定者を設定する
func (m *Mendikot) SetTrumpForTest(suit, chooser int) { m.trumpSuit, m.trumpChooserIdx = suit, chooser }

// SetDealerIdxForTest ディーラーを設定する
func (m *Mendikot) SetDealerIdxForTest(i int) { m.dealerIdx = i }

// SetWinnerTeamForTest 勝利チームを設定する
func (m *Mendikot) SetWinnerTeamForTest(team int) { m.winnerTeam = team }

// GiveTricksForTest 指定プレイヤーに空のトリックを n 個持たせる
func (m *Mendikot) GiveTricksForTest(playerIdx, n int) {
	for range n {
		m.players[playerIdx].AddTrick([]*Card{})
	}
}

// FinishHandForTest ハンド精算を直接呼ぶ
func (m *Mendikot) FinishHandForTest() { m.finishHand() }

// FinishGameForTest 終局処理を直接呼ぶ
func (m *Mendikot) FinishGameForTest() { m.finishGame() }

// PlayForTest 指定プレイヤーに 1 枚出させる
func (m *Mendikot) PlayForTest(playerIdx, cardIndex int) error { return m.play(playerIdx, cardIndex) }

// CpuChoiceForTest CPU が選ぶ手札インデックスを返す
func (m *Mendikot) CpuChoiceForTest(playerIdx int) int { return m.chooseCpuCard(playerIdx) }
