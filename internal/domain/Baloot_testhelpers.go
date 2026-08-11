//go:build test

package domain

// Baloot のテスト専用セッター。**配りは乱数なので、テストは盤面を直接組む。**
// 本番コードからは呼ばれない（`test` タグ付きでのみコンパイルされる）。

// SetPhaseForTest フェーズを設定する
func (b *Baloot) SetPhaseForTest(phase BalootPhase) { b.phase = phase }

// SetModeForTest モードを設定する
func (b *Baloot) SetModeForTest(mode BalootMode) { b.mode = mode }

// SetTrickNumberForTest トリック番号を設定する
func (b *Baloot) SetTrickNumberForTest(n int) { b.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する
func (b *Baloot) SetCurrentPlayerIdxForTest(i int) { b.currentPlayerIdx = i }

// SetCurrentTrickForTest 現在のトリックを設定する
func (b *Baloot) SetCurrentTrickForTest(tc []*TrickCard) { b.currentTrick = tc }

// SetTrumpSuitForTest 切り札のスートを設定する
func (b *Baloot) SetTrumpSuitForTest(suit int) { b.trumpSuit = suit }

// SetDeclarerIdxForTest 宣言者を設定する
func (b *Baloot) SetDeclarerIdxForTest(i int) { b.declarerIdx = i }

// SetRoundPointsForTest チームの現ラウンド点を設定する
func (b *Baloot) SetRoundPointsForTest(team, n int) {
	if team >= 0 && team < BalootTeamCnt {
		b.roundPoints[team] = n
	}
}

// SetDealerIdxForTest ディーラーを設定する
func (b *Baloot) SetDealerIdxForTest(i int) { b.dealerIdx = i }

// FinishGameForTest 終局処理を直接呼ぶ
func (b *Baloot) FinishGameForTest() { b.finishGame() }
