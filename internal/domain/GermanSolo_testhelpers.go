//go:build test

package domain

// GetCallableAceSuitsForTest はテスト用に席 idx が呼べるエースのスートを返す。
// GetCallableAceSuits と違いフェーズを問わない。
func (g *GermanSolo) GetCallableAceSuitsForTest(idx int) []int {
	return g.callableAceSuits(idx)
}

// SetPartnerForTest はテスト用に味方の席と公開状態を設定する。
func (g *GermanSolo) SetPartnerForTest(idx int, revealed bool) {
	g.partnerIdx = idx
	g.partnerRevealed = revealed
}

// SetPlaysAloneForTest はテスト用に単独プレイを設定する。
func (g *GermanSolo) SetPlaysAloneForTest(v bool) { g.playsAlone = v }

// SetCalledAceSuitForTest はテスト用に呼ばれたエースのスートを設定する。
func (g *GermanSolo) SetCalledAceSuitForTest(suit int) { g.calledAceSuit = suit }

// CpuBidForTest はテスト用に席 idx に 1 回ビッドさせる (手番を問わない)。
func (g *GermanSolo) CpuBidForTest(idx int) {
	bid := g.cpuChooseBid(idx)
	trump := -1
	if bid != GermanSoloBidNone {
		trump = g.cpuChooseTrump(idx)
	}
	g.bids[idx] = bid
	g.bidActed[idx] = true
	g.bidTrump[idx] = trump
}

// StartPlayForTest はテスト用にプレイフェーズを開始する。
func (g *GermanSolo) StartPlayForTest() { g.startPlay() }
