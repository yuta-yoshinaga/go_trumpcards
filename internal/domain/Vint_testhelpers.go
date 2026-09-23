//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを設定する。
func (v *Vint) SetPhaseForTest(p VintPhase) { v.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (v *Vint) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(v.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (v *Vint) SetContractForTest(declarer, level, denom int) {
	v.declarerIdx = declarer
	v.highBid = &VintBid{Player: declarer, Level: level, Denom: denom}
	v.trumpSuit = VintDenomToSuit(denom)
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (v *Vint) SetCurrentPlayerForTest(idx int) { v.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (v *Vint) SetTrickLeaderForTest(idx int) { v.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (v *Vint) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < VintPlayerCnt {
		v.tricksWon[idx] = n
	}
}

// SetTakenForTest はテスト用にチームが取った札を設定する。
func (v *Vint) SetTakenForTest(team int, cards []*Card) {
	if team >= 0 && team < VintTeamCnt {
		v.takenCards[team] = cards
	}
}

// SetBelowForTest はテスト用に線下の点を設定する。
func (v *Vint) SetBelowForTest(team, n int) {
	if team >= 0 && team < VintTeamCnt {
		v.below[team] = n
	}
}

// SetGamesWonForTest はテスト用に取ったゲーム数を設定する。
func (v *Vint) SetGamesWonForTest(team, n int) {
	if team >= 0 && team < VintTeamCnt {
		v.gamesWon[team] = n
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (v *Vint) FinishHandForTest() { v.finishHand() }
