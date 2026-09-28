//go:build test

package domain

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (k *Karnoffel) SetPhaseForTest(p KarnoffelPhase) { k.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (k *Karnoffel) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(k.GetPlayer(idx), cards)
}

// SetChosenSuitForTest は選ばれたスートを差し替える (テスト専用)。
func (k *Karnoffel) SetChosenSuitForTest(suit int) { k.chosenSuit = suit }

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (k *Karnoffel) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetDealerForTest は親を差し替える (テスト専用)。
func (k *Karnoffel) SetDealerForTest(idx int) { k.dealerIdx = idx }

// SetTrickLeaderForTest はリード席を差し替える (テスト専用)。
func (k *Karnoffel) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetTrickNumberForTest は済んだトリック数を差し替える (テスト専用)。
func (k *Karnoffel) SetTrickNumberForTest(n int) { k.trickNumber = n }

// SetTricksWonForTest は取得トリック数を差し替える (テスト専用)。
func (k *Karnoffel) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < KarnoffelPlayerCnt {
		k.tricksWon[idx] = n
	}
}

// SetHandsWonForTest は取得局数を差し替える (テスト専用)。
func (k *Karnoffel) SetHandsWonForTest(team, n int) {
	if team >= 0 && team < KarnoffelTeamCnt {
		k.handsWon[team] = n
	}
}

// FinishHandForTest は精算を走らせる (テスト専用)。
func (k *Karnoffel) FinishHandForTest() { k.finishHand() }
