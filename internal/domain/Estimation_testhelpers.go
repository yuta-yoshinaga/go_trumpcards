//go:build test

package domain

// Estimation のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (e *Estimation) SetPhaseForTest(phase EstimationPhase) { e.phase = phase }

// SetTrickNumberForTest トリック番号を設定する
func (e *Estimation) SetTrickNumberForTest(n int) { e.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (e *Estimation) SetCurrentPlayerIdxForTest(i int) { e.currentPlayerIdx = i }

// SetBidPlayerIdxForTest 宣言の手番を設定する
func (e *Estimation) SetBidPlayerIdxForTest(i int) { e.bidPlayerIdx = i }

// SetLeadPlayerIdxForTest リードプレイヤーを設定する
func (e *Estimation) SetLeadPlayerIdxForTest(i int) { e.leadPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (e *Estimation) SetCurrentTrickForTest(tc []*TrickCard) { e.currentTrick = tc }

// SetTrumpSuitForTest 切り札のスートを設定する
func (e *Estimation) SetTrumpSuitForTest(suit int) { e.trumpSuit = suit }

// SetDealerIdxForTest ディーラーを設定する。宣言の手番も併せて動かす。
func (e *Estimation) SetDealerIdxForTest(i int) {
	e.dealerIdx = i
	if e.phase == EstimationPhaseTrump {
		e.currentPlayerIdx = i
		e.bidPlayerIdx = i
	}
}

// SetWinnerIdxForTest 勝者を設定する
func (e *Estimation) SetWinnerIdxForTest(i int) { e.winnerIdx = i }

// SetBidsForTest 複数プレイヤーの宣言をまとめて設定する。0 は Dash 扱い。
func (e *Estimation) SetBidsForTest(bids map[int]int) {
	for idx, bid := range bids {
		if idx < 0 || idx >= len(e.players) {
			continue
		}
		e.players[idx].SetBid(bid)
		if bid == 0 {
			e.players[idx].SetCallType(EstimationCallDash)
		}
	}
}

// CloseBiddingForTest 宣言の締め切り処理を直接呼ぶ
func (e *Estimation) CloseBiddingForTest() { e.closeBidding() }

// FinishGameForTest 終局処理を直接呼ぶ
func (e *Estimation) FinishGameForTest() { e.finishGame() }

// PlayForTest 指定プレイヤーに 1 枚出させる
func (e *Estimation) PlayForTest(playerIdx, cardIndex int) error {
	return e.play(playerIdx, cardIndex)
}

// CpuChoiceForTest CPU が選ぶ手札インデックスを返す
func (e *Estimation) CpuChoiceForTest(playerIdx int) int { return e.chooseCpuCard(playerIdx) }

// EstimateForTest CPU の見積もりトリック数を返す
func (e *Estimation) EstimateForTest(playerIdx int) int { return e.estimateTricks(playerIdx) }
