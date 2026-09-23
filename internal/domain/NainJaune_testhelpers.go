//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (n *NainJaune) SetPhaseForTest(p NainJaunePhase) { n.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (n *NainJaune) SetCurrentPlayerForTest(idx int) { n.currentIdx = idx }

// SetBoardForTest はテスト用に盤を差し替える。
func (n *NainJaune) SetBoardForTest(b NainJauneBoard) { n.board = b }

// SetRunRankForTest はテスト用に並びの状態を差し替える。
func (n *NainJaune) SetRunRankForTest(rank int) { n.runRank = rank }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (n *NainJaune) SetDealNumberForTest(d int) { n.dealNo = d }
