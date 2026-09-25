package domain

import "encoding/json"

// BriscolaPlayer ブリスコラプレイヤークラス
type BriscolaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewBriscolaPlayer コンストラクタ
func NewBriscolaPlayer(isHuman bool) *BriscolaPlayer {
	return &BriscolaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲームをリセット (手札/トリック/上がり状態を初期化)
func (p *BriscolaPlayer) ResetGame() {
	resetPlayerRound(p)
}

// briscolaPlayerJSON is the JSON wire format for BriscolaPlayer.
type briscolaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *BriscolaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(briscolaPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BriscolaPlayer) UnmarshalJSON(data []byte) error {
	var j briscolaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.TrickHolder, j.TrickHolder)
	return nil
}
