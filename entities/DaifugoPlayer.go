package entities

import "sort"

// DaifugoPlayer 大富豪プレイヤークラス
type DaifugoPlayer struct {
	Player
	isHuman    bool
	isFinished bool
	rank       int // -1 = まだ確定していない, 1 = 大富豪, 2 = 富豪, 3 = 平民, 4 = 大貧民
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
func (p *DaifugoPlayer) GetIsHuman() bool { return p.isHuman }

// GetIsFinished 上がっているかどうか
func (p *DaifugoPlayer) GetIsFinished() bool { return p.isFinished }

// SetIsFinished 上がり状態設定
func (p *DaifugoPlayer) SetIsFinished(v bool) { p.isFinished = v }

// GetRank ランク取得 (-1 = 未確定)
func (p *DaifugoPlayer) GetRank() int { return p.rank }

// SetRank ランク設定
func (p *DaifugoPlayer) SetRank(r int) { p.rank = r }

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

// SortCards カードを大富豪のルールに従った強さ順 (弱い順) にソート
func (p *DaifugoPlayer) SortCards() {
	sort.Slice(p.cards, func(i, j int) bool {
		return DaifugoCardStrength(p.cards[i].GetValue()) < DaifugoCardStrength(p.cards[j].GetValue())
	})
}
