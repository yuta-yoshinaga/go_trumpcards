//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"sort"
)

// Tichu宣言種別
const (
	// TichuDeclNone 宣言なし
	TichuDeclNone = 0
	// TichuDeclTichu ティチュー宣言 (±100)
	TichuDeclTichu = 1
	// TichuDeclGrand グランドティチュー宣言 (±200)
	TichuDeclGrand = 2
)

// TichuPlayer ティチュープレイヤー
type TichuPlayer struct {
	*GamePlayer
	rank      int     // 上がり順位 (0=未上がり, 1=最初に上がり ...)
	declType  int     // 宣言種別 (TichuDeclNone/Tichu/Grand)
	collected []*Card // このディール中にトリックで獲得したカード (得点計算用)
}

// tichuPlayerJSON is the JSON wire format for TichuPlayer.
type tichuPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Rank       int         `json:"rk"`
	DeclType   int         `json:"dt"`
	Collected  []*Card     `json:"co"`
}

// MarshalJSON implements json.Marshaler.
func (p *TichuPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tichuPlayerJSON{
		GamePlayer: p.GamePlayer,
		Rank:       p.rank,
		DeclType:   p.declType,
		Collected:  p.collected,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TichuPlayer) UnmarshalJSON(data []byte) error {
	var j tichuPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.rank = j.Rank
	p.declType = j.DeclType
	p.collected = j.Collected
	if p.collected == nil {
		p.collected = make([]*Card, 0)
	}
	return nil
}

// NewTichuPlayer コンストラクタ
func NewTichuPlayer(isHuman bool) *TichuPlayer {
	return &TichuPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		rank:       0,
		declType:   TichuDeclNone,
		collected:  make([]*Card, 0),
	}
}

// GetRank 上がり順位取得 (0=未上がり)
func (p *TichuPlayer) GetRank() int { return p.rank }

// SetRank 上がり順位設定
func (p *TichuPlayer) SetRank(r int) { p.rank = r }

// GetDeclType 宣言種別取得
func (p *TichuPlayer) GetDeclType() int { return p.declType }

// SetDeclType 宣言種別設定
func (p *TichuPlayer) SetDeclType(d int) { p.declType = d }

// GetCollected 獲得カード取得
func (p *TichuPlayer) GetCollected() []*Card { return p.collected }

// AddCollected トリックで獲得したカードを追加
func (p *TichuPlayer) AddCollected(cards []*Card) {
	p.collected = append(p.collected, cards...)
}

// SortCardsByStrength カードをティチューの強さ順 (弱い順) にソート
func (p *TichuPlayer) SortCardsByStrength() {
	sort.Slice(p.cards, func(i, j int) bool {
		return TichuCardStrength(p.cards[i]) < TichuCardStrength(p.cards[j])
	})
}
