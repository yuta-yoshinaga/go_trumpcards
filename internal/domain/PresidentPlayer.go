package domain

import (
	"encoding/json"
	"sort"
)

// PresidentPlayer プレジデントのプレイヤー
type PresidentPlayer struct {
	*RankedGamePlayer
	prevRank int // 前回のランク (-1 = なし)
}

// presidentPlayerJSON is the JSON wire format for PresidentPlayer.
type presidentPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
	PrevRank         int               `json:"pr"`
}

// MarshalJSON implements json.Marshaler.
func (p *PresidentPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(presidentPlayerJSON{
		RankedGamePlayer: p.RankedGamePlayer,
		PrevRank:         p.prevRank,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PresidentPlayer) UnmarshalJSON(data []byte) error {
	var j presidentPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RankedGamePlayer != nil {
		p.RankedGamePlayer = j.RankedGamePlayer
	} else {
		p.RankedGamePlayer = NewRankedGamePlayer(false)
	}
	p.prevRank = j.PrevRank
	return nil
}

// NewPresidentPlayer constructs a PresidentPlayer.
func NewPresidentPlayer(isHuman bool) *PresidentPlayer {
	return &PresidentPlayer{
		RankedGamePlayer: NewRankedGamePlayer(isHuman),
		prevRank:         -1,
	}
}

// GetPrevRank 前回のランク取得 (-1 = なし)
func (p *PresidentPlayer) GetPrevRank() int { return p.prevRank }

// SetPrevRank 前回のランク設定
func (p *PresidentPlayer) SetPrevRank(r int) { p.prevRank = r }

// SortCardsByStrength カードを指定強さ関数で弱い順にソート
func (p *PresidentPlayer) SortCardsByStrength(strengthFn func(*Card) int) {
	sort.Slice(p.cards, func(i, j int) bool {
		return strengthFn(p.cards[i]) < strengthFn(p.cards[j])
	})
}
