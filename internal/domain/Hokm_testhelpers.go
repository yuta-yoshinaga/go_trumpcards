//go:build test

package domain

// Hokm のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (h *Hokm) SetPhaseForTest(phase HokmPhase) { h.phase = phase }

// SetTrickNumberForTest トリック番号を設定する
func (h *Hokm) SetTrickNumberForTest(n int) { h.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (h *Hokm) SetCurrentPlayerIdxForTest(i int) { h.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest リードプレイヤーを設定する
func (h *Hokm) SetLeadPlayerIdxForTest(i int) { h.leadPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (h *Hokm) SetCurrentTrickForTest(tc []*TrickCard) { h.currentTrick = tc }

// SetTrumpSuitForTest 切り札のスートを設定する
func (h *Hokm) SetTrumpSuitForTest(suit int) { h.trumpSuit = suit }

// SetHakemIdxForTest 親を設定する。配り直後なら手番も併せて動かす。
func (h *Hokm) SetHakemIdxForTest(i int) {
	h.hakemIdx = i
	if h.phase == HokmPhaseTrump {
		h.currentPlayerIdx = i
		h.leadPlayerIdx = i
	}
}

// SetWinnerTeamForTest 勝利チームを設定する
func (h *Hokm) SetWinnerTeamForTest(team int) { h.winnerTeam = team }

// GiveTricksForTest 指定プレイヤーに空のトリックを n 個持たせる
func (h *Hokm) GiveTricksForTest(playerIdx, n int) {
	for range n {
		h.players[playerIdx].AddTrick([]*Card{})
	}
}

// FinishHandForTest ハンド精算を直接呼ぶ
func (h *Hokm) FinishHandForTest(winnerTeam int) { h.finishHand(winnerTeam) }

// FinishGameForTest 終局処理を直接呼ぶ
func (h *Hokm) FinishGameForTest() { h.finishGame() }

// PlayForTest 指定プレイヤーに 1 枚出させる
func (h *Hokm) PlayForTest(playerIdx, cardIndex int) error { return h.play(playerIdx, cardIndex) }

// CpuChoiceForTest CPU が選ぶ手札インデックスを返す
func (h *Hokm) CpuChoiceForTest(playerIdx int) int { return h.chooseCpuCard(playerIdx) }
