//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを設定する。
func (k *Klaberjass) SetPhaseForTest(p KlaberjassPhase) { k.phase = p }

// SetTrumpForTest はテスト用に切札を設定する。
func (k *Klaberjass) SetTrumpForTest(suit int) { k.trumpSuit = suit }

// SetMakerForTest はテスト用にメイカーを設定する。
func (k *Klaberjass) SetMakerForTest(idx int) { k.makerIdx = idx }

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (k *Klaberjass) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (k *Klaberjass) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetHandForTest はテスト用に手札を差し替える。
func (k *Klaberjass) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(k.GetPlayer(idx), cards)
}

// SetHandPointsForTest はテスト用にディール中の得点を設定する。
func (k *Klaberjass) SetHandPointsForTest(idx, pts int) {
	if idx >= 0 && idx < KlaberjassPlayerCnt {
		k.handPoints[idx] = pts
	}
}

// SetTurnUpForTest はテスト用に表向きカードを設定する。
func (k *Klaberjass) SetTurnUpForTest(c *Card) { k.turnUpCard = c }

// SetScoreForTest はテスト用に通算点を設定する。
func (k *Klaberjass) SetScoreForTest(idx, score int) {
	if idx >= 0 && idx < KlaberjassPlayerCnt {
		k.scores[idx] = score
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (k *Klaberjass) FinishHandForTest() { k.finishHand() }

// CollectSequencesForTest はテスト用に役の比較を走らせる。
func (k *Klaberjass) CollectSequencesForTest() { k.collectSequences() }

// FindBelaForTest はテスト用にベラ保持者を探す。
func (k *Klaberjass) FindBelaForTest() { k.findBela() }
