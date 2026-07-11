//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"sort"
)

// ZhengPlayer 争上游プレイヤークラス
type ZhengPlayer struct {
	*RankedGamePlayer
}

// zhengPlayerJSON is the JSON wire format for ZhengPlayer.
type zhengPlayerJSON struct {
	RankedGamePlayer *RankedGamePlayer `json:"rp"`
}

// MarshalJSON implements json.Marshaler.
func (p *ZhengPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(zhengPlayerJSON{
		RankedGamePlayer: p.RankedGamePlayer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ZhengPlayer) UnmarshalJSON(data []byte) error {
	var j zhengPlayerJSON
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

// NewZhengPlayer コンストラクタ
func NewZhengPlayer(isHuman bool) *ZhengPlayer {
	return &ZhengPlayer{
		RankedGamePlayer: NewRankedGamePlayer(isHuman),
	}
}

// SortCardsByZhengStrength 争上游のランク順(弱→強)にソートする (同ランクはスート順で安定表示)
func (p *ZhengPlayer) SortCardsByZhengStrength() {
	sort.Slice(p.cards, func(i, j int) bool {
		si, sj := zhengRankStrength(p.cards[i]), zhengRankStrength(p.cards[j])
		if si != sj {
			return si < sj
		}
		return p.cards[i].GetDesign() < p.cards[j].GetDesign()
	})
}
