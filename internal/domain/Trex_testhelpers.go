//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (t *Trex) SetPhaseForTest(p TrexPhase) { t.phase = p }

// SetContractForTest はテスト用に契約を差し替える。
func (t *Trex) SetContractForTest(c TrexContract) { t.contract = c }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (t *Trex) SetCurrentPlayerForTest(idx int) { t.currentIdx = idx }

// SetKingForTest はテスト用に王を差し替える。
func (t *Trex) SetKingForTest(idx int) { t.kingIdx = idx }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (t *Trex) SetDealNumberForTest(n int) { t.dealNo = n }
