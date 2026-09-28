//go:build test

package domain

// SetLayoutForTest はテスト用に場札を差し替える。
func (l *LaughAndLieDown) SetLayoutForTest(cards []*Card) { l.layout = cards }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (l *LaughAndLieDown) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }
