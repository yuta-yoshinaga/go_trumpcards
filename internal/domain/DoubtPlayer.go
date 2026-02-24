package domain

import "sort"

// DoubtPlayer ダウトプレイヤークラス
type DoubtPlayer struct {
	Player
	isHuman    bool
	isFinished bool
}

// NewDoubtPlayer コンストラクタ
func NewDoubtPlayer(isHuman bool) *DoubtPlayer {
	return &DoubtPlayer{
		Player:     Player{cards: make([]*Card, 0)},
		isHuman:    isHuman,
		isFinished: false,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *DoubtPlayer) GetIsHuman() bool { return p.isHuman }

// GetIsFinished 上がり済みかどうか
func (p *DoubtPlayer) GetIsFinished() bool { return p.isFinished }

// SetIsFinished 上がり状態設定
func (p *DoubtPlayer) SetIsFinished(v bool) { p.isFinished = v }

// RemoveCards 指定インデックスのカードを手札から取り除いて返す
// 重複インデックスは無視する。返却カードは元のインデックス昇順。
func (p *DoubtPlayer) RemoveCards(indices []int) []*Card {
	if len(indices) == 0 {
		return []*Card{}
	}

	// 重複除去
	seen := make(map[int]bool)
	deduped := make([]int, 0, len(indices))
	for _, idx := range indices {
		if !seen[idx] {
			seen[idx] = true
			deduped = append(deduped, idx)
		}
	}

	// 降順ソートして後ろから削除 (インデックスずれを防ぐ)
	sort.Sort(sort.Reverse(sort.IntSlice(deduped)))

	// 降順で削除し、取り除いたカードを記録 (降順のまま格納)
	removed := make([]*Card, 0, len(deduped))
	for _, idx := range deduped {
		if idx < 0 || idx >= len(p.cards) {
			continue
		}
		removed = append(removed, p.cards[idx])
		p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
	}

	// 降順で格納したので逆順にして昇順に戻す
	for i, j := 0, len(removed)-1; i < j; i, j = i+1, j-1 {
		removed[i], removed[j] = removed[j], removed[i]
	}
	return removed
}
