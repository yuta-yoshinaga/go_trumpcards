//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを設定する。
func (k *Kaiser) SetPhaseForTest(p KaiserPhase) { k.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (k *Kaiser) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(k.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (k *Kaiser) SetContractForTest(declarer, value, suit int, contract KaiserContract) {
	k.declarerIdx = declarer
	k.highBid = &KaiserBid{Player: declarer, Value: value, Contract: contract}
	k.trumpSuit = suit
	k.contract = contract
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (k *Kaiser) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (k *Kaiser) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetHandPointsForTest はテスト用に局中の得点を設定する。
func (k *Kaiser) SetHandPointsForTest(team, pts int) {
	if team >= 0 && team < KaiserTeamCnt {
		k.handPoints[team] = pts
	}
}

// SetScoreForTest はテスト用に通算点を設定する。
func (k *Kaiser) SetScoreForTest(team, score int) {
	if team >= 0 && team < KaiserTeamCnt {
		k.scores[team] = score
	}
}

// SetTrickNumberForTest はテスト用に済んだトリック数を設定する。
func (k *Kaiser) SetTrickNumberForTest(n int) { k.trickNumber = n }

// FinishHandForTest はテスト用に精算を走らせる。
func (k *Kaiser) FinishHandForTest() { k.finishHand() }
