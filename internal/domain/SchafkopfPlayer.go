//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// SchafkopfPlayer シャーフコップのプレイヤークラス。手札（GamePlayer）、獲得
// トリック（TrickHolder）、所持チップ（ChipHolder）を保持する。チーム（ピッカー
// 組／ディフェンダー組）は Schafkopf 本体がラウンドごとに管理する。
type SchafkopfPlayer struct {
	*GamePlayer
	TrickHolder
	ChipHolder
}

// NewSchafkopfPlayer コンストラクタ
func NewSchafkopfPlayer(isHuman bool, startChips int) *SchafkopfPlayer {
	p := &SchafkopfPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
	p.SetChips(startChips)
	return p
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）。チップは
// 累積するためリセットしない。
func (p *SchafkopfPlayer) ResetRound() {
	resetPlayerRound(p)
}

// schafkopfPlayerJSON is the JSON wire format for SchafkopfPlayer.
type schafkopfPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	ChipHolder  *ChipHolder  `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *SchafkopfPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(schafkopfPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		ChipHolder:  &p.ChipHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SchafkopfPlayer) UnmarshalJSON(data []byte) error {
	var j schafkopfPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	if j.ChipHolder != nil {
		p.ChipHolder = *j.ChipHolder
	}
	return nil
}
