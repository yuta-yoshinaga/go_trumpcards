package domain

import (
	"math/rand"
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
