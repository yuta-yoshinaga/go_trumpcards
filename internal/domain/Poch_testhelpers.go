//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (p *Poch) SetPhaseForTest(ph PochPhase) { p.phase = ph }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (p *Poch) SetCurrentPlayerForTest(idx int) { p.currentIdx = idx }

// SetPaySuitForTest はテスト用に pay suit を差し替える。
func (p *Poch) SetPaySuitForTest(suit int) { p.paySuit = suit }

// SetBoardForTest はテスト用に盤を差し替える。
func (p *Poch) SetBoardForTest(b PochBoard) { p.board = b }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (p *Poch) SetDealNumberForTest(n int) { p.dealNo = n }

// SetStopsForTest はテスト用に並びの状態を差し替える。
func (p *Poch) SetStopsForTest(suit, rank int) { p.stopsSuit, p.stopsRank = suit, rank }

// ResolveStakingForTest はテスト用に第 1 段階だけを走らせる。
func (p *Poch) ResolveStakingForTest() { p.resolveStaking() }
