//go:build test

package domain

// SetPhase テスト用: フェーズ設定
func (d *Doudizhu) SetPhase(phase DoudizhuPhase) { d.round.phase = phase }

// SetCurrentTurn テスト用: 手番設定
func (d *Doudizhu) SetCurrentTurn(turn int) { d.round.currentTurn = turn }

// SetTableCombo テスト用: 場の役設定
func (d *Doudizhu) SetTableCombo(combo *DoudizhuCombo) { d.round.tableCombo = combo }

// SetLastPlayIdx テスト用: 最後に出したプレイヤー設定
func (d *Doudizhu) SetLastPlayIdx(idx int) { d.round.lastPlayIdx = idx }

// SetGameEndFlag テスト用: ゲーム終了フラグ設定
func (d *Doudizhu) SetGameEndFlag(flag bool) { d.round.gameEndFlag = flag }

// SetKittyCards テスト用: 底牌設定
func (d *Doudizhu) SetKittyCards(cards []*Card) { d.round.kittyCards = cards }

// SetLandlordIdx テスト用: 地主インデックス設定
func (d *Doudizhu) SetLandlordIdx(idx int) {
	d.round.landlordIdx = idx
	for i, p := range d.players {
		p.SetIsLandlord(i == idx)
	}
}

// SetBaseBid テスト用: ビッド値設定
func (d *Doudizhu) SetBaseBid(bid int) { d.round.baseBid = bid }

// SetBombCount テスト用: ボムカウント設定
func (d *Doudizhu) SetBombCount(count int) { d.round.bombCount = count }

// SetPassCount テスト用: パスカウント設定
func (d *Doudizhu) SetPassCount(count int) { d.round.passCount = count }

// SetHighestBid テスト用: 最高ビッド設定
func (d *Doudizhu) SetHighestBid(bid int) { d.round.highestBid = bid }

// SetHighestBidder テスト用: 最高入札者設定
func (d *Doudizhu) SetHighestBidder(idx int) { d.round.highestBidder = idx }
