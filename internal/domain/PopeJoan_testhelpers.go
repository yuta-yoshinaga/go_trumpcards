//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (p *PopeJoan) SetPhaseForTest(ph PopeJoanPhase) { p.phase = ph }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (p *PopeJoan) SetCurrentPlayerForTest(idx int) { p.currentIdx = idx }

// SetTrumpSuitForTest はテスト用にトランプを差し替える。
func (p *PopeJoan) SetTrumpSuitForTest(suit int) { p.trumpSuit = suit }

// SetBoardForTest はテスト用に盤を差し替える。
func (p *PopeJoan) SetBoardForTest(b PopeJoanBoard) { p.board = b }

// SetRunForTest はテスト用に並びの状態を差し替える。
func (p *PopeJoan) SetRunForTest(suit, rank int) { p.runSuit, p.runRank = suit, rank }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (p *PopeJoan) SetDealNumberForTest(n int) { p.dealNo = n }

// SetTurnUpForTest はテスト用にめくり札を差し替える。
func (p *PopeJoan) SetTurnUpForTest(c *Card) { p.turnUp = c }

// ResolveTurnUpForTest はテスト用にめくり札の精算だけを走らせる。
func (p *PopeJoan) ResolveTurnUpForTest() { p.resolveTurnUp() }
