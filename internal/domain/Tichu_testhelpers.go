//go:build test

package domain

// SetScores テスト用: チーム得点設定
func (t *Tichu) SetScores(scores [2]int) { t.round.scores = scores }

// SetGameEndFlag テスト用: 終了フラグ設定
func (t *Tichu) SetGameEndFlag(flag bool) { t.round.gameEndFlag = flag }
