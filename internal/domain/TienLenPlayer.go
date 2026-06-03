//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"sort"
)

// TienLenPlayer Tien Lenプレイヤークラス
type TienLenPlayer struct {
	*RankedGamePlayer
}

// tienLenPlayerJSON is the JSON wire format for TienLenPlayer.
type tienLenPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
}

// MarshalJSON implements json.Marshaler.
func (p *TienLenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tienLenPlayerJSON{
		RankedGamePlayer: p.RankedGamePlayer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TienLenPlayer) UnmarshalJSON(data []byte) error {
	var j tienLenPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RankedGamePlayer != nil {
		p.RankedGamePlayer = j.RankedGamePlayer
	} else {
		p.RankedGamePlayer = NewRankedGamePlayer(false)
	}
	return nil
}

// NewTienLenPlayer コンストラクタ
func NewTienLenPlayer(isHuman bool) *TienLenPlayer {
	return &TienLenPlayer{
		RankedGamePlayer: NewRankedGamePlayer(isHuman),
	}
}

// SortCardsByTienLenStrength Tien Lenの強さ順(弱→強)にソートする
func (p *TienLenPlayer) SortCardsByTienLenStrength() {
	sort.Slice(p.cards, func(i, j int) bool {
		return TienLenCardStrength(p.cards[i]) < TienLenCardStrength(p.cards[j])
	})
}
