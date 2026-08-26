//go:build !js || !wasm || extra

package domain

// SpeculationPlayer はスペキュレーションの 1 人分の状態。
//
// **座席 0 が人間**で、以降が CPU。座席番号はラウンドを跨いで変わらない。
type SpeculationPlayer struct {
	// name は表示名。
	name string
	// chips は所持チップ。
	chips ChipHolder
	// hidden はまだめくっていない伏せ札。先頭からめくる。
	hidden []*Card
	// won は買い取った / 自分でめくって手元に残っている切り札のうち、
	// **現在の最高札**。持っていなければ nil。
	//
	// 競りで売った札はここから移る。「誰が最高の切り札を持っているか」が
	// このゲームの勝敗そのものなので、札そのものを持たせる。
	best *Card
}

// NewSpeculationPlayer は名前と初期チップを与えてプレイヤーを作る。
func NewSpeculationPlayer(name string, chips int) *SpeculationPlayer {
	p := &SpeculationPlayer{name: name}
	p.chips.SetChips(chips)
	return p
}

// GetName は表示名を返す。
func (p *SpeculationPlayer) GetName() string { return p.name }

// GetChips は所持チップを返す。
func (p *SpeculationPlayer) GetChips() int { return p.chips.GetChips() }

// AddChips はチップを加える。
func (p *SpeculationPlayer) AddChips(n int) { p.chips.AddChips(n) }

// SubtractChips はチップを引く。足りなければ false を返して何もしない。
func (p *SpeculationPlayer) SubtractChips(n int) bool { return p.chips.SubtractChips(n) }

// SetChips はチップを設定する（テスト・復元用）。
func (p *SpeculationPlayer) SetChips(n int) { p.chips.SetChips(n) }

// GetHiddenCount はまだめくっていない伏せ札の枚数を返す。
func (p *SpeculationPlayer) GetHiddenCount() int { return len(p.hidden) }

// GetBest は現在持っている最高の切り札を返す。持っていなければ nil。
func (p *SpeculationPlayer) GetBest() *Card { return p.best }

// SetBest は最高の切り札を設定する（買い取り・売却・復元で使う）。
func (p *SpeculationPlayer) SetBest(c *Card) { p.best = c }

// SetHidden は伏せ札を設定する（配布・復元・テスト用）。
func (p *SpeculationPlayer) SetHidden(cards []*Card) { p.hidden = cards }

// GetHidden は伏せ札を返す（永続化・テスト用）。
func (p *SpeculationPlayer) GetHidden() []*Card { return p.hidden }

// FlipTop は伏せ札の先頭を 1 枚めくって返す。無ければ nil。
func (p *SpeculationPlayer) FlipTop() *Card {
	if len(p.hidden) == 0 {
		return nil
	}
	c := p.hidden[0]
	p.hidden = p.hidden[1:]
	return c
}
