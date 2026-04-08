package domain

import (
	"encoding/json"
	"sort"
)

// DurakPlayer ドゥラークプレイヤークラス
type DurakPlayer struct {
	*GamePlayer
}

// durakPlayerJSON is the JSON wire format for DurakPlayer.
type durakPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *DurakPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(durakPlayerJSON{
		GamePlayer: p.GamePlayer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DurakPlayer) UnmarshalJSON(data []byte) error {
	var j durakPlayerJSON
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

// NewDurakPlayer コンストラクタ
func NewDurakPlayer(isHuman bool) *DurakPlayer {
	return &DurakPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// SortCards カードをスート・値順にソート
func (p *DurakPlayer) SortCards(trumpSuit int) {
	sort.Slice(p.cards, func(i, j int) bool {
		ci, cj := p.cards[i], p.cards[j]
		isTrumpI := ci.GetDesign() == trumpSuit
		isTrumpJ := cj.GetDesign() == trumpSuit
		// 非切り札を先、切り札を後ろ
		if isTrumpI != isTrumpJ {
			return !isTrumpI
		}
		// 同じグループ内ではスート昇順
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		// 同スートでは値昇順 (Aは最大=14扱い)
		return durakCardRank(ci) < durakCardRank(cj)
	})
}

// SortCardsByValue カードを値順にソート (スート無視)
func (p *DurakPlayer) SortCardsByValue(trumpSuit int) {
	sort.Slice(p.cards, func(i, j int) bool {
		ci, cj := p.cards[i], p.cards[j]
		ri, rj := durakCardRank(ci), durakCardRank(cj)
		if ri != rj {
			return ri < rj
		}
		// 同ランクではスート順 (切り札を後ろに)
		isTrumpI := ci.GetDesign() == trumpSuit
		isTrumpJ := cj.GetDesign() == trumpSuit
		if isTrumpI != isTrumpJ {
			return !isTrumpI
		}
		return ci.GetDesign() < cj.GetDesign()
	})
}
