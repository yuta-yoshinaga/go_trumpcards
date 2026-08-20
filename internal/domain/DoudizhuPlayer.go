//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"sort"
)

// DoudizhuPlayer 斗地主プレイヤー
type DoudizhuPlayer struct {
	*GamePlayer
	isLandlord bool
}

// doudizhuPlayerJSON is the JSON wire format for DoudizhuPlayer.
type doudizhuPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	IsLandlord bool        `json:"ll"`
}

// MarshalJSON implements json.Marshaler.
func (p *DoudizhuPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(doudizhuPlayerJSON{
		GamePlayer: p.GamePlayer,
		IsLandlord: p.isLandlord,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DoudizhuPlayer) UnmarshalJSON(data []byte) error {
	var j doudizhuPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.isLandlord = j.IsLandlord
	return nil
}

// NewDoudizhuPlayer コンストラクタ
func NewDoudizhuPlayer(isHuman bool) *DoudizhuPlayer {
	return &DoudizhuPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		isLandlord: false,
	}
}

// GetIsLandlord 地主かどうか
func (p *DoudizhuPlayer) GetIsLandlord() bool { return p.isLandlord }

// SetIsLandlord 地主設定
func (p *DoudizhuPlayer) SetIsLandlord(v bool) { p.isLandlord = v }

// SortCardsByStrength カードを斗地主の強さ順 (弱い順) にソート
func (p *DoudizhuPlayer) SortCardsByStrength() {
	sort.Slice(p.cards, func(i, j int) bool {
		return DoudizhuCardStrength(p.cards[i]) < DoudizhuCardStrength(p.cards[j])
	})
}
