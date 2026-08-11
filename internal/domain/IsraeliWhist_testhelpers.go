//go:build test

package domain

// IsraeliWhist のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (w *IsraeliWhist) SetPhaseForTest(phase IsraeliWhistPhase) { w.phase = phase }

// SetTrickNumberForTest トリック番号を設定する
func (w *IsraeliWhist) SetTrickNumberForTest(n int) { w.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (w *IsraeliWhist) SetCurrentPlayerIdxForTest(i int) { w.currentPlayerIdx = i }

// SetAuctionPlayerIdxForTest オークションの手番を設定する
func (w *IsraeliWhist) SetAuctionPlayerIdxForTest(i int) { w.auctionPlayerIdx = i }

// SetBidPlayerIdxForTest 宣言の手番を設定する
func (w *IsraeliWhist) SetBidPlayerIdxForTest(i int) { w.bidPlayerIdx = i }

// SetLeadPlayerIdxForTest リードプレイヤーを設定する
func (w *IsraeliWhist) SetLeadPlayerIdxForTest(i int) { w.leadPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (w *IsraeliWhist) SetCurrentTrickForTest(tc []*TrickCard) { w.currentTrick = tc }

// SetTrumpSuitForTest 切り札のスートを設定する
func (w *IsraeliWhist) SetTrumpSuitForTest(suit int) { w.trumpSuit = suit }

// SetDealerIdxForTest ディーラーを設定する。配り直後なら両段の手番も動かす。
func (w *IsraeliWhist) SetDealerIdxForTest(i int) {
	w.dealerIdx = i
	if w.phase == IsraeliWhistPhaseAuction {
		w.auctionPlayerIdx = (i + 1) % IsraeliWhistPlayerCnt
		w.bidPlayerIdx = w.auctionPlayerIdx
		w.currentPlayerIdx = w.auctionPlayerIdx
	}
}

// SetDeclarerForTest 落札者と最低ノルマを設定する
func (w *IsraeliWhist) SetDeclarerForTest(idx, bid, suit int) {
	w.declarerIdx, w.highBid, w.highSuit = idx, bid, suit
}

// SetWinnerIdxForTest 勝者を設定する
func (w *IsraeliWhist) SetWinnerIdxForTest(i int) { w.winnerIdx = i }

// SetBidsForTest 複数プレイヤーの 2 段階目の宣言をまとめて設定する
func (w *IsraeliWhist) SetBidsForTest(bids map[int]int) {
	for idx, bid := range bids {
		if idx >= 0 && idx < len(w.players) {
			w.players[idx].SetBid(bid)
		}
	}
}

// CloseAuctionForTest オークションの締め切り処理を直接呼ぶ
func (w *IsraeliWhist) CloseAuctionForTest() { w.closeAuction() }

// FinishGameForTest 終局処理を直接呼ぶ
func (w *IsraeliWhist) FinishGameForTest() { w.finishGame() }

// PlayForTest 指定プレイヤーに 1 枚出させる
func (w *IsraeliWhist) PlayForTest(playerIdx, cardIndex int) error {
	return w.play(playerIdx, cardIndex)
}

// CpuChoiceForTest CPU が選ぶ手札インデックスを返す
func (w *IsraeliWhist) CpuChoiceForTest(playerIdx int) int { return w.chooseCpuCard(playerIdx) }

// OutbidsForTest 入札が現在の最高入札を上回るかを返す
func (w *IsraeliWhist) OutbidsForTest(bid, suit int) bool { return w.outbids(bid, suit) }

// FinishRoundForTest ラウンド精算を直接呼ぶ
func (w *IsraeliWhist) FinishRoundForTest() { w.finishRound() }
