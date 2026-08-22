//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// PutPlayer プットプレイヤークラス。
// 手札 (GamePlayer) と当該マノで獲得したバサ (トリック) を保持する。
type PutPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewPutPlayer コンストラクタ
func NewPutPlayer(isHuman bool) *PutPlayer {
	return &PutPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame 手札とバサ獲得状態を初期化する (新しいマノの配り直し用)。
func (p *PutPlayer) ResetGame() {
	resetPlayerRound(p)
}

// putPlayerJSON is the JSON wire format for PutPlayer.
type putPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *PutPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(putPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PutPlayer) UnmarshalJSON(data []byte) error {
	var j putPlayerJSON
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
	return nil
}
