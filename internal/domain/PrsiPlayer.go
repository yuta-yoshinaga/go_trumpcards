package domain

import "encoding/json"

// PrsiPlayer プルシープレイヤークラス
type PrsiPlayer struct {
	*GamePlayer
}

// NewPrsiPlayer コンストラクタ
func NewPrsiPlayer(isHuman bool) *PrsiPlayer {
	return &PrsiPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// prsiPlayerJSON is the JSON wire format for PrsiPlayer.
type prsiPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *PrsiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(prsiPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PrsiPlayer) UnmarshalJSON(data []byte) error {
	var j prsiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	return nil
}
