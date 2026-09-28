//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを設定する。
func (b *BidEuchre) SetPhaseForTest(p BidEuchrePhase) { b.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (b *BidEuchre) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(b.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (b *BidEuchre) SetContractForTest(declarer, value int, t BidEuchreTrump) {
	b.declarerIdx = declarer
	b.highBid = &BidEuchreBid{Player: declarer, Value: value}
	b.trump = t
	b.trumpSuit = BidEuchreTrumpSuit(t)
	b.trumpChosen = true
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (b *BidEuchre) SetCurrentPlayerForTest(idx int) { b.currentIdx = idx }

// SetDealerForTest はテスト用にディーラーを設定する。
func (b *BidEuchre) SetDealerForTest(idx int) { b.dealerIdx = idx }

// SetBidPlayerForTest はテスト用に宣言手番を設定する。
func (b *BidEuchre) SetBidPlayerForTest(idx int) { b.bidIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (b *BidEuchre) SetTrickLeaderForTest(idx int) { b.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (b *BidEuchre) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < BidEuchrePlayerCnt {
		b.tricksWon[idx] = n
	}
}

// SetScoreForTest はテスト用に通算点を設定する。
func (b *BidEuchre) SetScoreForTest(team, n int) {
	if team >= 0 && team < BidEuchreTeamCnt {
		b.scores[team] = n
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (b *BidEuchre) FinishHandForTest() { b.finishHand() }
