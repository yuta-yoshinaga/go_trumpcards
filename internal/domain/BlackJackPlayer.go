package domain

import "encoding/json"

// BlackJackPlayer ブラックジャックプレイヤークラス
type BlackJackPlayer struct {
	Player     // 親クラス
	ChipHolder // チップ管理
}

// NewBlackJackPlayer コンストラクタ
func NewBlackJackPlayer() *BlackJackPlayer {
	return &BlackJackPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
	}
}

// GetScore 手札から現在のスコア計算
func (bp *BlackJackPlayer) GetScore() int {
	return CalculateBlackJackScore(bp.cards)
}

// IsSoft ソフトハンド（11として有効なエースを含む）かどうか判定
func (bp *BlackJackPlayer) IsSoft() bool {
	_, isSoft := calcScore(bp.cards)
	return isSoft
}

// blackJackPlayerJSON is the JSON wire format for BlackJackPlayer.
type blackJackPlayerJSON struct {
	Player     *Player     `json:"p"`
	ChipHolder *ChipHolder `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (bp *BlackJackPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackJackPlayerJSON{
		Player:     &bp.Player,
		ChipHolder: &bp.ChipHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (bp *BlackJackPlayer) UnmarshalJSON(data []byte) error {
	var j blackJackPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		bp.Player = *j.Player
	}
	if j.ChipHolder != nil {
		bp.ChipHolder = *j.ChipHolder
	}
	return nil
}
