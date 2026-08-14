//go:build !js || !wasm || classic

package domain

import "encoding/json"

// BotifarraPlayer はボティファラのプレイヤー。
//
// 手札 (GamePlayer) と獲得したトリック (TrickHolder) を持ちます。**点はチーム単位**
// なので、ここには持たせず本体が集計します。
type BotifarraPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewBotifarraPlayer はコンストラクタ。
func NewBotifarraPlayer(isHuman bool) *BotifarraPlayer {
	return &BotifarraPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound はラウンドの状態に戻す。
func (p *BotifarraPlayer) ResetRound() {
	resetPlayerRound(p)
}

// botifarraPlayerJSON is the JSON wire format for BotifarraPlayer.
type botifarraPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *BotifarraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(botifarraPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BotifarraPlayer) UnmarshalJSON(data []byte) error {
	var j botifarraPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errBotifarraMissingBase
	}
	p.GamePlayer = j.GamePlayer
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	return nil
}
