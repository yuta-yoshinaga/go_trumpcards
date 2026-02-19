package entities

import "sort"

// DaifugoPlayer 大富豪プレイヤークラス
type DaifugoPlayer struct {
	Player
	isHuman    bool
	isFinished bool
	rank       int // 0:大富豪, 1:富豪, 2:平民, 3:貧民, 4:大貧民 (初期値-1)
}

// NewDaifugoPlayer コンストラクタ
func NewDaifugoPlayer(isHuman bool) *DaifugoPlayer {
	return &DaifugoPlayer{
		Player:     Player{cards: make([]*Card, 0)},
		isHuman:    isHuman,
		isFinished: false,
		rank:       -1,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *DaifugoPlayer) GetIsHuman() bool {
	return p.isHuman
}

// GetIsFinished 上がっているかどうか
func (p *DaifugoPlayer) GetIsFinished() bool {
	return p.isFinished
}

// SetIsFinished 上がり状態設定
func (p *DaifugoPlayer) SetIsFinished(v bool) {
	p.isFinished = v
}

// GetRank ランク取得
func (p *DaifugoPlayer) GetRank() int {
	return p.rank
}

// SetRank ランク設定
func (p *DaifugoPlayer) SetRank(v int) {
	p.rank = v
}

// SortCards 手札をソート (大富豪の強さ順)
func (p *DaifugoPlayer) SortCards() {
	sort.Slice(p.cards, func(i, j int) bool {
		return GetDaifugoStrength(p.cards[i]) < GetDaifugoStrength(p.cards[j])
	})
}

// GetDaifugoStrength カードの強さを取得 (3=0, ..., 2=12, Joker=13)
func GetDaifugoStrength(c *Card) int {
	if c.GetDesign() == CardDesignJoker {
		return 13
	}
	val := c.GetValue()
	if val == 1 { // A
		return 11
	}
	if val == 2 { // 2
		return 12
	}
	return val - 3 // 3=0, 4=1, ..., 13(K)=10
}

// RemoveCards 指定したインデックスのカードを手札から取り除く
func (p *DaifugoPlayer) RemoveCards(indices []int) []*Card {
	removed := make([]*Card, 0, len(indices))
	// インデックスを降順にソートして、後ろから取り除く
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	for _, idx := range indices {
		if idx < 0 || idx >= len(p.cards) {
			continue
		}
		removed = append(removed, p.cards[idx])
		p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
	}
	return removed
}
