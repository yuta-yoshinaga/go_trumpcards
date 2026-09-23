//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを設定する。
func (b *Boston) SetPhaseForTest(p BostonPhase) { b.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (b *Boston) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(b.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (b *Boston) SetContractForTest(declarer, partner int, level BostonBidLevel, suit int) {
	b.declarerIdx = declarer
	b.partnerIdx = partner
	b.highBid = &BostonBidRecord{Player: declarer, Level: level, Suit: suit}
	b.trumpSuit = suit
	b.exposed = BostonBidIsExposed(level)
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (b *Boston) SetCurrentPlayerForTest(idx int) { b.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (b *Boston) SetTrickLeaderForTest(idx int) { b.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (b *Boston) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < BostonPlayerCnt {
		b.tricksWon[idx] = n
	}
}

// SetHandNumberForTest はテスト用に局番号を設定する。
func (b *Boston) SetHandNumberForTest(n int) { b.handNumber = n }

// FinishHandForTest はテスト用に精算を走らせる。
func (b *Boston) FinishHandForTest() { b.finishHand() }
