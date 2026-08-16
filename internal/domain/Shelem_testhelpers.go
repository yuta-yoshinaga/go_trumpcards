//go:build test

package domain

// Shelem のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (s *Shelem) SetPhaseForTest(phase ShelemPhase) { s.phase = phase }

// SetTrickNumberForTest トリック番号を設定する
func (s *Shelem) SetTrickNumberForTest(n int) { s.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (s *Shelem) SetCurrentPlayerIdxForTest(i int) { s.currentPlayerIdx = i }

// SetBidPlayerIdxForTest 競りの手番を設定する
func (s *Shelem) SetBidPlayerIdxForTest(i int) { s.bidPlayerIdx = i }

// SetLeadPlayerIdxForTest リードプレイヤーを設定する
func (s *Shelem) SetLeadPlayerIdxForTest(i int) { s.leadPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (s *Shelem) SetCurrentTrickForTest(tc []*TrickCard) { s.currentTrick = tc }

// SetTrumpSuitForTest 切り札のスートを設定する
func (s *Shelem) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetDealerIdxForTest ディーラーを設定する。配り直後なら競りの手番も動かす。
func (s *Shelem) SetDealerIdxForTest(i int) {
	s.dealerIdx = i
	if s.phase == ShelemPhaseBid {
		s.bidPlayerIdx = (i + 1) % ShelemPlayerCnt
		s.currentPlayerIdx = s.bidPlayerIdx
	}
}

// SetContractForTest 落札者・契約点・Shelem 宣言を設定する
func (s *Shelem) SetContractForTest(idx, contract int, shelem bool) {
	s.declarerIdx, s.contract, s.shelemBid = idx, contract, shelem
}

// SetRoundPointsForTest チームの現ラウンドのカード点を設定する
func (s *Shelem) SetRoundPointsForTest(team, n int) {
	if team >= 0 && team < ShelemTeamCnt {
		s.roundPoints[team] = n
	}
}

// SetWinnerTeamForTest 勝利チームを設定する
func (s *Shelem) SetWinnerTeamForTest(team int) { s.winnerTeam = team }

// GiveTricksForTest 指定プレイヤーに空のトリックを n 個持たせる
func (s *Shelem) GiveTricksForTest(playerIdx, n int) {
	for range n {
		s.players[playerIdx].AddTrick([]*Card{})
	}
}

// CloseBiddingForTest 競りの締め切り処理を直接呼ぶ
func (s *Shelem) CloseBiddingForTest() { s.closeBidding() }

// FinishRoundForTest ラウンド精算を直接呼ぶ
func (s *Shelem) FinishRoundForTest() { s.finishRound() }

// FinishGameForTest 終局処理を直接呼ぶ
func (s *Shelem) FinishGameForTest() { s.finishGame() }

// PlayForTest 指定プレイヤーに 1 枚出させる
func (s *Shelem) PlayForTest(playerIdx, cardIndex int) error { return s.play(playerIdx, cardIndex) }

// CpuChoiceForTest CPU が選ぶ手札インデックスを返す
func (s *Shelem) CpuChoiceForTest(playerIdx int) int { return s.chooseCpuCard(playerIdx) }

// CpuDiscardAndTrumpForTest CPU 落札者の捨て札・切り札決定を直接呼ぶ
func (s *Shelem) CpuDiscardAndTrumpForTest() { s.cpuDiscardAndTrump() }
