package domain

import (
	"encoding/json"
	"sort"
)

// BigTwoPlayer Big Twoプレイヤークラス
type BigTwoPlayer struct {
	*RankedGamePlayer
}

// bigTwoPlayerJSON is the JSON wire format for BigTwoPlayer.
type bigTwoPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BigTwoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigTwoPlayerJSON{
		RankedGamePlayer: p.RankedGamePlayer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BigTwoPlayer) UnmarshalJSON(data []byte) error {
	var j bigTwoPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RankedGamePlayer != nil {
		p.RankedGamePlayer = j.RankedGamePlayer
	} else {
		p.RankedGamePlayer = NewRankedGamePlayer(false)
	}
	return nil
}

// NewBigTwoPlayer コンストラクタ
func NewBigTwoPlayer(isHuman bool) *BigTwoPlayer {
	return &BigTwoPlayer{
		RankedGamePlayer: NewRankedGamePlayer(isHuman),
	}
}

// SortCardsByBigTwoStrength Big Twoの強さ順(弱→強)にソートする
func (p *BigTwoPlayer) SortCardsByBigTwoStrength() {
	sort.Slice(p.cards, func(i, j int) bool {
		si := BigTwoCardStrength(p.cards[i])
		sj := BigTwoCardStrength(p.cards[j])
		return si < sj
	})
}
