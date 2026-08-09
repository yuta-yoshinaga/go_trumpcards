package domain

import "encoding/json"

// TwoTenJackPlayer ツーテンジャックプレイヤークラス
type TwoTenJackPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewTwoTenJackPlayer コンストラクタ
func NewTwoTenJackPlayer(isHuman bool) *TwoTenJackPlayer {
	return &TwoTenJackPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット
func (p *TwoTenJackPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// GetCapturedPointCards 獲得した点札の合計得点 (A=1, 10=10, J=1)
func (p *TwoTenJackPlayer) GetCapturedPointCards() int {
	total := 0
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			total += TwoTenJackCardPoints(c)
		}
	}
	return total
}

// TwoTenJackCardPoints カード1枚の得点を返す (A=1, 10=10, J=1, その他=0)
func TwoTenJackCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 1
	case 10:
		return 10
	case 11:
		return 1
	default:
		return 0
	}
}

// twoTenJackPlayerJSON is the JSON wire format for TwoTenJackPlayer.
type twoTenJackPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TwoTenJackPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(twoTenJackPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TwoTenJackPlayer) UnmarshalJSON(data []byte) error {
	var j twoTenJackPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	return nil
}
