package domain

import (
	"encoding/json"
	"sort"
)

// DaifugoPlayer 大富豪プレイヤークラス
type DaifugoPlayer struct {
	*RankedGamePlayer
	prevRank             int  // 前回のランク (-1 = なし)
	illegalFinishPenalty bool // 反則上がりペナルティ
}

// daifugoPlayerJSON is the JSON wire format for DaifugoPlayer.
type daifugoPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
	PrevRank         int               `json:"pr"`
	IllegalFinish    bool              `json:"if"`
}

// MarshalJSON implements json.Marshaler.
func (p *DaifugoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(daifugoPlayerJSON{
		RankedGamePlayer: p.RankedGamePlayer,
		PrevRank:         p.prevRank,
		IllegalFinish:    p.illegalFinishPenalty,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DaifugoPlayer) UnmarshalJSON(data []byte) error {
	var j daifugoPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.RankedGamePlayer != nil {
		p.RankedGamePlayer = j.RankedGamePlayer
	} else {
		p.RankedGamePlayer = NewRankedGamePlayer(false)
	}
	p.prevRank = j.PrevRank
	p.illegalFinishPenalty = j.IllegalFinish
	return nil
}

// NewDaifugoPlayer コンストラクタ
func NewDaifugoPlayer(isHuman bool) *DaifugoPlayer {
	return &DaifugoPlayer{
		RankedGamePlayer:     NewRankedGamePlayer(isHuman),
		prevRank:             -1,
		illegalFinishPenalty: false,
	}
}

// GetPrevRank 前回のランク取得 (-1 = なし)
func (p *DaifugoPlayer) GetPrevRank() int { return p.prevRank }

// SetPrevRank 前回のランク設定
func (p *DaifugoPlayer) SetPrevRank(r int) { p.prevRank = r }

// GetIllegalFinishPenalty 反則上がりペナルティ取得
func (p *DaifugoPlayer) GetIllegalFinishPenalty() bool { return p.illegalFinishPenalty }

// SetIllegalFinishPenalty 反則上がりペナルティ設定
func (p *DaifugoPlayer) SetIllegalFinishPenalty(v bool) { p.illegalFinishPenalty = v }

// SortCardsByStrength カードを指定強さ関数で弱い順にソート
func (p *DaifugoPlayer) SortCardsByStrength(strengthFn func(*Card) int) {
	sort.Slice(p.cards, func(i, j int) bool {
		return strengthFn(p.cards[i]) < strengthFn(p.cards[j])
	})
}

// SortCards カードを大富豪の通常ルールに従った強さ順 (弱い順) にソート
func (p *DaifugoPlayer) SortCards() {
	p.SortCardsByStrength(func(c *Card) int {
		if IsJoker(c) {
			return DaifugoJokerStrength
		}
		return DaifugoCardStrength(c.GetValue())
	})
}

// RemoveCardsByValue 指定した値のカードを全て除去して返す
func (p *DaifugoPlayer) RemoveCardsByValue(value int) []*Card {
	removed := make([]*Card, 0)
	remaining := make([]*Card, 0)
	for _, c := range p.cards {
		if !IsJoker(c) && c.GetValue() == value {
			removed = append(removed, c)
		} else {
			remaining = append(remaining, c)
		}
	}
	p.cards = remaining
	return removed
}

// CountCardsByValue 指定した値のカードの枚数を返す
func (p *DaifugoPlayer) CountCardsByValue(value int) int {
	count := 0
	for _, c := range p.cards {
		if !IsJoker(c) && c.GetValue() == value {
			count++
		}
	}
	return count
}
