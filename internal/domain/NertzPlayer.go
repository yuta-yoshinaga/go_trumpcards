package domain

import "encoding/json"

// Nertz のプレイヤー盤面構造定数
const (
	// NertzTableauCnt 1プレイヤーあたりのタブロー(リバー)列数
	NertzTableauCnt = 4
	// NertzPileSize ナッツパイルの初期枚数
	NertzPileSize = 13
	// NertzInitialStockSize 配り終わり時点のストック枚数 (52 - 13 - 4)
	NertzInitialStockSize = 35
)

// NertzTableauCard タブロー上のカード。Klondike とほぼ同形だが、
// Nertz では基本的に常に表向き (FaceUp=true) で扱う。テスト/将来拡張用に
// FaceUp フィールドを保持する。
type NertzTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// NertzPlayer Nertz のプレイヤー
type NertzPlayer struct {
	name      string
	isCpu     bool
	deckIdx   int
	nertzPile []*Card // 末尾 (最後の要素) が表向きトップ
	tableau   [NertzTableauCnt][]*NertzTableauCard
	stock     []*Card // 末尾がトップ (次に引かれるカード)
	waste     []*Card // 末尾がトップ
	score     int     // 累積スコア (ラウンド跨ぎで保持)
}

// NewNertzPlayer コンストラクタ
func NewNertzPlayer(name string, isCpu bool, deckIdx int) *NertzPlayer {
	return &NertzPlayer{
		name:    name,
		isCpu:   isCpu,
		deckIdx: deckIdx,
	}
}

// GetName プレイヤー名
func (p *NertzPlayer) GetName() string { return p.name }

// GetIsCpu CPU かどうか
func (p *NertzPlayer) GetIsCpu() bool { return p.isCpu }

// GetDeckIdx 自分のデッキインデックス (= ファウンデーション帰属識別)
func (p *NertzPlayer) GetDeckIdx() int { return p.deckIdx }

// GetScore 累積スコア
func (p *NertzPlayer) GetScore() int { return p.score }

// SetScore スコア設定 (テスト/復元用)
func (p *NertzPlayer) SetScore(s int) { p.score = s }

// AddScore スコア加算 (負値も可)
func (p *NertzPlayer) AddScore(delta int) { p.score += delta }

// ResetRoundPiles ラウンド開始時にすべての盤面パイルをクリアする。
// 累積スコアは保持される。
func (p *NertzPlayer) ResetRoundPiles() {
	p.nertzPile = nil
	p.stock = nil
	p.waste = nil
	for i := range p.tableau {
		p.tableau[i] = nil
	}
}

// --- Nertz pile ---

// NertzTop ナッツパイルのトップ (次にプレイされるカード)。空なら nil。
func (p *NertzPlayer) NertzTop() *Card {
	if len(p.nertzPile) == 0 {
		return nil
	}
	return p.nertzPile[len(p.nertzPile)-1]
}

// PopNertz ナッツパイルのトップを取り出す
func (p *NertzPlayer) PopNertz() *Card {
	if len(p.nertzPile) == 0 {
		return nil
	}
	c := p.nertzPile[len(p.nertzPile)-1]
	p.nertzPile = p.nertzPile[:len(p.nertzPile)-1]
	return c
}

// PushNertz ナッツパイルにカードを積む (テスト/配札用)
func (p *NertzPlayer) PushNertz(c *Card) {
	p.nertzPile = append(p.nertzPile, c)
}

// NertzSize ナッツパイルの残り枚数
func (p *NertzPlayer) NertzSize() int { return len(p.nertzPile) }

// GetNertzPile ナッツパイルのスナップショット (末尾がトップ)
func (p *NertzPlayer) GetNertzPile() []*Card {
	out := make([]*Card, len(p.nertzPile))
	copy(out, p.nertzPile)
	return out
}

// --- Tableau ---

// TableauSize タブロー列 i の枚数
func (p *NertzPlayer) TableauSize(i int) int {
	if i < 0 || i >= NertzTableauCnt {
		return 0
	}
	return len(p.tableau[i])
}

// TableauTop タブロー列 i の最下段 (= 末尾) のカード本体。空または範囲外なら nil。
func (p *NertzPlayer) TableauTop(i int) *Card {
	if i < 0 || i >= NertzTableauCnt {
		return nil
	}
	col := p.tableau[i]
	if len(col) == 0 {
		return nil
	}
	return col[len(col)-1].Card
}

// GetTableauColumn タブロー列 i のスナップショット (末尾がトップ = 最下段)
func (p *NertzPlayer) GetTableauColumn(i int) []*NertzTableauCard {
	if i < 0 || i >= NertzTableauCnt {
		return nil
	}
	out := make([]*NertzTableauCard, len(p.tableau[i]))
	copy(out, p.tableau[i])
	return out
}

// PushTableau タブロー列 i にカードを積む。範囲外は無視。
func (p *NertzPlayer) PushTableau(i int, tc *NertzTableauCard) {
	if i < 0 || i >= NertzTableauCnt || tc == nil {
		return
	}
	p.tableau[i] = append(p.tableau[i], tc)
}

// TakeTableauTail タブロー列 i の fromIdx 以降を切り出して返す。元の列はそこで切り詰められる。
// 範囲外は nil を返し、何も変更しない。
func (p *NertzPlayer) TakeTableauTail(i, fromIdx int) []*NertzTableauCard {
	if i < 0 || i >= NertzTableauCnt {
		return nil
	}
	col := p.tableau[i]
	if fromIdx < 0 || fromIdx >= len(col) {
		return nil
	}
	tail := make([]*NertzTableauCard, len(col)-fromIdx)
	copy(tail, col[fromIdx:])
	p.tableau[i] = col[:fromIdx]
	return tail
}

// --- Waste ---

// WasteTop ウェイストのトップ
func (p *NertzPlayer) WasteTop() *Card {
	if len(p.waste) == 0 {
		return nil
	}
	return p.waste[len(p.waste)-1]
}

// PopWaste ウェイストのトップを取り出す
func (p *NertzPlayer) PopWaste() *Card {
	if len(p.waste) == 0 {
		return nil
	}
	c := p.waste[len(p.waste)-1]
	p.waste = p.waste[:len(p.waste)-1]
	return c
}

// PushWaste ウェイストに積む
func (p *NertzPlayer) PushWaste(c *Card) { p.waste = append(p.waste, c) }

// WasteSize ウェイスト枚数
func (p *NertzPlayer) WasteSize() int { return len(p.waste) }

// GetWaste ウェイストのスナップショット (末尾がトップ)
func (p *NertzPlayer) GetWaste() []*Card {
	out := make([]*Card, len(p.waste))
	copy(out, p.waste)
	return out
}

// --- Stock ---

// StockSize ストック枚数
func (p *NertzPlayer) StockSize() int { return len(p.stock) }

// PushStock ストックの末尾に積む (= トップになる)
func (p *NertzPlayer) PushStock(c *Card) { p.stock = append(p.stock, c) }

// PopStock ストックのトップを取り出す
func (p *NertzPlayer) PopStock() *Card {
	if len(p.stock) == 0 {
		return nil
	}
	c := p.stock[len(p.stock)-1]
	p.stock = p.stock[:len(p.stock)-1]
	return c
}

// RecycleWasteToStock ストックが空のとき、ウェイスト全体をひっくり返してストックに戻す。
// 元の引き順を保つため、ウェイスト末尾 (=最後に置かれたカード) がストック末尾になるように
// 逆順で詰め直す: stock[0] が最も古いウェイストカード = 次に Pop される。
//
// Note: PopStock は末尾を返すため、リサイクル後に最初に Pop したいのは
// ウェイスト先頭 (一番下) のカード。よってウェイスト先頭がストック末尾に来るように
// 単純に逆順詰めする。
func (p *NertzPlayer) RecycleWasteToStock() {
	if len(p.waste) == 0 {
		return
	}
	// ウェイスト末尾 → ストック先頭、ウェイスト先頭 → ストック末尾
	stock := make([]*Card, len(p.waste))
	for i, c := range p.waste {
		stock[len(p.waste)-1-i] = c
	}
	p.stock = append(p.stock, stock...)
	p.waste = nil
}

// --- JSON ---

// nertzPlayerJSON is the JSON wire format for NertzPlayer.
type nertzPlayerJSON struct {
	Name      string                               `json:"nm"`
	IsCpu     bool                                 `json:"ic"`
	DeckIdx   int                                  `json:"di"`
	NertzPile []*Card                              `json:"np"`
	Tableau   [NertzTableauCnt][]*NertzTableauCard `json:"tb"`
	Stock     []*Card                              `json:"st"`
	Waste     []*Card                              `json:"wa"`
	Score     int                                  `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *NertzPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(nertzPlayerJSON{
		Name:      p.name,
		IsCpu:     p.isCpu,
		DeckIdx:   p.deckIdx,
		NertzPile: p.nertzPile,
		Tableau:   p.tableau,
		Stock:     p.stock,
		Waste:     p.waste,
		Score:     p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *NertzPlayer) UnmarshalJSON(data []byte) error {
	var j nertzPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.name = j.Name
	p.isCpu = j.IsCpu
	p.deckIdx = j.DeckIdx
	p.nertzPile = j.NertzPile
	p.tableau = j.Tableau
	for i := range p.tableau {
		if p.tableau[i] == nil {
			p.tableau[i] = make([]*NertzTableauCard, 0)
		}
	}
	p.stock = j.Stock
	p.waste = j.Waste
	p.score = j.Score
	return nil
}
