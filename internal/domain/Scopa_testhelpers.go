//go:build test

package domain

// ScopaTestNew は決定的なテスト用 Scopa を生成する (2 人, シャッフルなし)。
// 山札・場札・手札はテスト側で setter を使って構築する。
func ScopaTestNew(config ScopaConfig) *Scopa {
	players := make([]*ScopaPlayer, ScopaPlayerCnt)
	players[0] = NewScopaPlayer(true)
	for i := 1; i < ScopaPlayerCnt; i++ {
		players[i] = NewScopaPlayer(false)
	}
	return NewScopa(NewTrumpCardsScopa(), players, config)
}

// ScopaTestSetTable はテスト用に場札を直接設定する。
func (s *Scopa) ScopaTestSetTable(cards []*Card) { s.round.tableCards = cards }

// ScopaTestSetPhase はテスト用にフェーズを設定する。
func (s *Scopa) ScopaTestSetPhase(phase string) { s.round.phase = phase }

// ScopaTestSetCurrentTurn はテスト用に手番を設定する。
func (s *Scopa) ScopaTestSetCurrentTurn(idx int) { s.round.currentTurn = idx }

// ScopaTestSetLastCapture はテスト用に最後の捕獲者を設定する。
func (s *Scopa) ScopaTestSetLastCapture(idx int) { s.round.lastCaptureIdx = idx }

// ScopaTestSetGameEnd はテスト用にゲーム終了フラグを設定する。
func (s *Scopa) ScopaTestSetGameEnd(flag bool) { s.round.gameEndFlag = flag }

// ScopaTestDrainDeck はテスト用に山札を空にする。
func (s *Scopa) ScopaTestDrainDeck() {
	for s.trumpCards.GetRemainingCount() > 0 {
		s.trumpCards.DrawCard()
	}
}

// ScopaTestFinishRound はテスト用にラウンド終了処理を直接呼ぶ。
func (s *Scopa) ScopaTestFinishRound() { s.finishRound() }

// ScopaTestScoreRound はテスト用に得点計算を直接呼ぶ。
func (s *Scopa) ScopaTestScoreRound() *ScopaScoreDetail { return s.scoreRound() }
