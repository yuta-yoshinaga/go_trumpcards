//go:build test

package domain

// SetScores テスト用: チーム得点設定
func (t *Tichu) SetScores(scores [2]int) { t.round.scores = scores }

// SetGameEndFlag テスト用: 終了フラグ設定
func (t *Tichu) SetGameEndFlag(flag bool) { t.round.gameEndFlag = flag }

// SetBombCountForTest はテスト用にボム使用回数を設定する。
func (t *Tichu) SetBombCountForTest(n int) { t.round.bombCount = n }

// SetIsOneTwoForTest はテスト用にワンツー成立を設定する。
func (t *Tichu) SetIsOneTwoForTest(v bool) { t.round.oneTwo = v }

// SetPhaseForTest はテスト用にフェーズを設定する。
func (t *Tichu) SetPhaseForTest(phase TichuPhase) { t.round.phase = phase }

// SetCurrentTurnForTest はテスト用に手番を設定する。
func (t *Tichu) SetCurrentTurnForTest(idx int) { t.round.currentTurn = idx }

// SetDogLeadPassedForTest はテスト用に犬リードフラグを設定する。
func (t *Tichu) SetDogLeadPassedForTest(passed bool, from int) {
	t.round.dogLeadPassed = passed
	t.round.dogLeadFrom = from
}
