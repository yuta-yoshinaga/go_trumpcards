//go:build test

package domain

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (l *Literature) SetPhaseForTest(p LiteraturePhase) { l.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (l *Literature) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(l.GetPlayer(idx), cards)
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (l *Literature) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }

// SetHalfSuitForTest はハーフスートの帰属を差し替える (テスト専用)。
func (l *Literature) SetHalfSuitForTest(half int, st LiteratureHalfSuitState) {
	if half >= 0 && half < LiteratureHalfSuitCnt {
		l.halfSuits[half] = st
	}
}

// CheckGameEndForTest は決着判定を走らせる (テスト専用)。
func (l *Literature) CheckGameEndForTest() { l.checkGameEnd() }
