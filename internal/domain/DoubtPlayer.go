package domain

import (
	"math/rand"
)

// cardMemoryEntry 記憶したカード1枚分の情報
type cardMemoryEntry struct {
	value    int // カードの値 (1-13)
	turnSeen int // 記憶したターン番号
}

// GetTurnSeen MemoryEntryインターフェース実装
func (e cardMemoryEntry) GetTurnSeen() int { return e.turnSeen }

// DoubtPlayer ダウトプレイヤークラス
type DoubtPlayer struct {
	*GamePlayer
	memoryManager[cardMemoryEntry]
}

// NewDoubtPlayer コンストラクタ
func NewDoubtPlayer(isHuman bool) *DoubtPlayer {
	return &DoubtPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// RecordRevealedCard 公開されたカードを記憶する (retentionChance の確率で記録)
func (p *DoubtPlayer) RecordRevealedCard(value int, retentionChance float64, turnNumber int) {
	if rand.Float64() < retentionChance {
		p.AddMemory(cardMemoryEntry{value: value, turnSeen: turnNumber})
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
