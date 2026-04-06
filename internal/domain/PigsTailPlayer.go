package domain

import "encoding/json"

// PigsTailPlayer ぶたのしっぽプレイヤークラス
type PigsTailPlayer struct {
	*GamePlayer
}

// pigsTailPlayerJSON is the JSON wire format for PigsTailPlayer.
type pigsTailPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *PigsTailPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigsTailPlayerJSON{
		GamePlayer: p.GamePlayer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PigsTailPlayer) UnmarshalJSON(data []byte) error {
	var j pigsTailPlayerJSON
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

// NewPigsTailPlayer コンストラクタ
func NewPigsTailPlayer(isHuman bool) *PigsTailPlayer {
	return &PigsTailPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}
