package domain

import (
	"math/rand"
	"sort"
)

// cardMemoryEntry 記憶したカード1枚分の情報
type cardMemoryEntry struct {
	value    int // カードの値 (1-13)
	turnSeen int // 記憶したターン番号
}

// DoubtPlayer ダウトプレイヤークラス
type DoubtPlayer struct {
	Player
	isHuman      bool
	isFinished   bool
	cardMemories []cardMemoryEntry // 記憶したカードのリスト
}

// NewDoubtPlayer コンストラクタ
func NewDoubtPlayer(isHuman bool) *DoubtPlayer {
	return &DoubtPlayer{
		Player:       Player{cards: make([]*Card, 0)},
		isHuman:      isHuman,
		isFinished:   false,
		cardMemories: nil,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *DoubtPlayer) GetIsHuman() bool { return p.isHuman }

// GetIsFinished 上がり済みかどうか
func (p *DoubtPlayer) GetIsFinished() bool { return p.isFinished }

// SetIsFinished 上がり状態設定
func (p *DoubtPlayer) SetIsFinished(v bool) { p.isFinished = v }

// ResetMemory カード記憶をリセットする
func (p *DoubtPlayer) ResetMemory() {
	p.cardMemories = nil
}

// RecordRevealedCard 公開されたカードを記憶する (retentionChance の確率で記録)
func (p *DoubtPlayer) RecordRevealedCard(value int, retentionChance float64, turnNumber int) {
	if rand.Float64() < retentionChance {
		p.cardMemories = append(p.cardMemories, cardMemoryEntry{value: value, turnSeen: turnNumber})
	}
}

// CountKnownCards 指定した値のカードを何枚知っているか返す (記憶 + 手札)
func (p *DoubtPlayer) CountKnownCards(value int) int {
	count := 0
	for _, entry := range p.cardMemories {
		if entry.value == value {
			count++
		}
	}
	for _, card := range p.cards {
		if card.GetValue() == value {
			count++
		}
	}
	return count
}

// DecayMemories 古い記憶を確率的に忘却する
// 忘却確率 = decayRate * (currentTurn - turnSeen)
func (p *DoubtPlayer) DecayMemories(currentTurn int, decayRate float64) {
	kept := p.cardMemories[:0]
	for _, entry := range p.cardMemories {
		age := currentTurn - entry.turnSeen
		forgetProb := decayRate * float64(age)
		if forgetProb >= 1.0 || rand.Float64() < forgetProb {
			continue // 忘却
		}
		kept = append(kept, entry)
	}
	p.cardMemories = kept
}

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
