//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// KempsPlayer はケムプスのプレイヤー。
//
// 手札 (GamePlayer 由来の cards) のみを保持する軽量プレイヤー。得点はチーム単位で
// ゲーム本体が管理するため、プレイヤー自身は得点を持たない。
type KempsPlayer struct {
	*GamePlayer
}

// NewKempsPlayer はコンストラクタ。
func NewKempsPlayer(isHuman bool) *KempsPlayer {
	return &KempsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// HasFourOfAKind は手札 4 枚がすべて同じランクかどうかを返す。
// 手札が KempsHandSize 枚未満のときは false。
func (p *KempsPlayer) HasFourOfAKind() bool {
	if p.GetCardsSize() < KempsHandSize {
		return false
	}
	first := p.GetCard(0)
	if first == nil {
		return false
	}
	v := first.GetValue()
	count := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c != nil && c.GetValue() == v {
			count++
		}
	}
	return count >= KempsHandSize
}

// kempsPlayerJSON is the JSON wire format for KempsPlayer.
type kempsPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *KempsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(kempsPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KempsPlayer) UnmarshalJSON(data []byte) error {
	var j kempsPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	return nil
}
