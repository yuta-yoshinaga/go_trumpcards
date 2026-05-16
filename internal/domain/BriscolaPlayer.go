package domain

import "encoding/json"

// BriscolaPlayerCnt ブリスコラのプレイヤー数 (v1は2人固定)
const BriscolaPlayerCnt = 2

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
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
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
