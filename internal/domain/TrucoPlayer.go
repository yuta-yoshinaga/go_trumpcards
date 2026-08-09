package domain

import "encoding/json"

// TrucoPlayer トゥルコプレイヤークラス。
// 手札 (GamePlayer) と当該マノで獲得したバサ (トリック) を保持する。
type TrucoPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTrucoPlayer コンストラクタ
func NewTrucoPlayer(isHuman bool) *TrucoPlayer {
	return &TrucoPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame 手札とバサ獲得状態を初期化する (新しいマノの配り直し用)。
func (p *TrucoPlayer) ResetGame() {
	resetPlayerRound(p)
}

// trucoPlayerJSON is the JSON wire format for TrucoPlayer.
type trucoPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TrucoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(trucoPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TrucoPlayer) UnmarshalJSON(data []byte) error {
	var j trucoPlayerJSON
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
