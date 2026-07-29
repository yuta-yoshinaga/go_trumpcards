//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BuraPlayer ブラのプレイヤークラス
type BuraPlayer struct {
	*GamePlayer
}

// NewBuraPlayer コンストラクタ
func NewBuraPlayer(isHuman bool) *BuraPlayer {
	return &BuraPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲームをリセット (手札と上がり状態を初期化)
func (p *BuraPlayer) ResetGame() {
	p.Reset()
	p.SetIsFinished(false)
}

// buraPlayerJSON is the JSON wire format for BuraPlayer.
type buraPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BuraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(buraPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BuraPlayer) UnmarshalJSON(data []byte) error {
	var j buraPlayerJSON
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
