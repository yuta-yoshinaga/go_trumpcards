//go:build test

package domain

// SetPhase テスト用: フェーズを設定する
func (d *Durak) SetPhase(phase DurakPhase) { d.round.phase = phase }

// SetAttackerIdx テスト用: 攻撃者インデックスを設定する
func (d *Durak) SetAttackerIdx(idx int) { d.round.attackerIdx = idx }

// SetDefenderIdx テスト用: 防御者インデックスを設定する
func (d *Durak) SetDefenderIdx(idx int) { d.round.defenderIdx = idx }

// SetCurrentTurn テスト用: 現在のターンを設定する
func (d *Durak) SetCurrentTurn(idx int) { d.round.currentTurn = idx }

// SetGameEndFlag テスト用: ゲーム終了フラグを設定する
func (d *Durak) SetGameEndFlag(v bool) { d.round.gameEndFlag = v }

// SetLoserIdx テスト用: 敗者インデックスを設定する
func (d *Durak) SetLoserIdx(idx int) { d.round.loserIdx = idx }

// SetTablePairs テスト用: テーブルペアを設定する
func (d *Durak) SetTablePairs(pairs []*DurakTablePair) { d.round.tablePairs = pairs }

// SetTrumpSuit テスト用: 切り札スートを設定する
func (d *Durak) SetTrumpSuit(suit int) { d.trumpSuit = suit }

// SetTrumpCard テスト用: 切り札カードを設定する
func (d *Durak) SetTrumpCard(card *Card) { d.trumpCard = card }

// SetStock テスト用: 山札を設定する
func (d *Durak) SetStock(stock []*Card) { d.stock = stock }

// SetBoutNumber テスト用: バウト番号を設定する
func (d *Durak) SetBoutNumber(n int) { d.round.boutNumber = n }

// GetStock テスト用: 山札を取得する
func (d *Durak) GetStock() []*Card { return d.stock }

// GetDiscardPile テスト用: 捨て札を取得する
func (d *Durak) GetDiscardPile() []*Card { return d.discardPile }

// SetDiscardPile テスト用: 捨て札を設定する
func (d *Durak) SetDiscardPile(pile []*Card) { d.discardPile = pile }
