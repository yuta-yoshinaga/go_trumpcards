//go:build test

package domain

// SetDeadForTest は死に手を差し替える (テスト用)。
func (c *Comet) SetDeadForTest(cards []*Card) { c.dead = cards }

// SetNeedForTest は次に要るランクを差し替える (テスト用)。
func (c *Comet) SetNeedForTest(n int) { c.need = n }

// SetPileForTest は連なりを差し替える (テスト用)。
func (c *Comet) SetPileForTest(cards []*Card) { c.pile = cards }

// SetCurrentForTest は手番の席を差し替える (テスト用)。
func (c *Comet) SetCurrentForTest(i int) { c.currentIdx = i }

// SetPhaseForTest はフェーズを差し替える (テスト用)。
func (c *Comet) SetPhaseForTest(p string) { c.phase = p }

// SetLastResultForTest は直前の局の集計を差し替える (テスト用)。
func (c *Comet) SetLastResultForTest(r *CometRoundResult) { c.lastResult = r }
