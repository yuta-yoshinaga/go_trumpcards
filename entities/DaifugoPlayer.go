package entities

import "sort"

// DaifugoPlayer 大富豪プレイヤークラス
type DaifugoPlayer struct {
	Player
	isHuman    bool
	isFinished bool
	rank       int // -1 = まだ確定していない, 1 = 大富豪, 2 = 富豪, 3 = 平民, 4 = 大貧民
	prevRank   int // 前回のランク (-1 = なし)
}

// NewDaifugoPlayer コンストラクタ
func NewDaifugoPlayer(isHuman bool) *DaifugoPlayer {
	return &DaifugoPlayer{
		Player:     Player{cards: make([]*Card, 0)},
		isHuman:    isHuman,
		isFinished: false,
		rank:       -1,
		prevRank:   -1,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *DaifugoPlayer) GetIsHuman() bool { return p.isHuman }

// GetIsFinished 上がっているかどうか
func (p *DaifugoPlayer) GetIsFinished() bool { return p.isFinished }

// SetIsFinished 上がり状態設定
func (p *DaifugoPlayer) SetIsFinished(v bool) { p.isFinished = v }

// GetRank ランク取得 (-1 = 未確定)
func (p *DaifugoPlayer) GetRank() int { return p.rank }

// SetRank ランク設定
func (p *DaifugoPlayer) SetRank(r int) { p.rank = r }

// GetPrevRank 前回のランク取得 (-1 = なし)
func (p *DaifugoPlayer) GetPrevRank() int { return p.prevRank }

// SetPrevRank 前回のランク設定
func (p *DaifugoPlayer) SetPrevRank(r int) { p.prevRank = r }

// RemoveCards 指定インデックスのカードを手札から取り除いて返す (昇順ソートされた順で返す)
// 重複するインデックスは無視される。
func (p *DaifugoPlayer) RemoveCards(indices []int) []*Card {
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Ints(sorted)

	// 重複排除
	unique := make([]int, 0, len(sorted))
	for i, idx := range sorted {
		if i == 0 || idx != sorted[i-1] {
			unique = append(unique, idx)
		}
	}

	removed := make([]*Card, len(unique))
	for i, idx := range unique {
		removed[i] = p.cards[idx]
	}

	// 後ろから削除してインデックスのズレを防ぐ
	for i := len(unique) - 1; i >= 0; i-- {
		idx := unique[i]
		p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
	}
	return removed
}

// SortCardsByStrength カードを指定強さ関数で弱い順にソート
func (p *DaifugoPlayer) SortCardsByStrength(strengthFn func(*Card) int) {
	sort.Slice(p.cards, func(i, j int) bool {
		return strengthFn(p.cards[i]) < strengthFn(p.cards[j])
	})
}

// SortCards カードを大富豪の通常ルールに従った強さ順 (弱い順) にソート
func (p *DaifugoPlayer) SortCards() {
	p.SortCardsByStrength(func(c *Card) int {
		if c.GetDesign() == CardDesignJoker {
			return DaifugoJokerStrength
		}
		return DaifugoCardStrength(c.GetValue())
	})
}
