//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (g *Goofspiel) SetPhaseForTest(p GoofspielPhase) { g.phase = p }

// SetCurrentPrizeForTest はいま公開されている賞札を設定する
func (g *Goofspiel) SetCurrentPrizeForTest(c *Card) { g.currentPrize = c }

// SetPrizePileForTest は未公開の賞札を設定する
func (g *Goofspiel) SetPrizePileForTest(cards []*Card) { g.prizePile = cards }

// BidForTest は指定席に入札させる
func (g *Goofspiel) BidForTest(playerIdx, cardIndex int) error { return g.bid(playerIdx, cardIndex) }

// ResolveForTest は入札を公開して解決する
func (g *Goofspiel) ResolveForTest() { g.resolve() }

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (g *Goofspiel) CpuChoiceForTest(playerIdx int) int { return g.chooseCpuCard(playerIdx) }
