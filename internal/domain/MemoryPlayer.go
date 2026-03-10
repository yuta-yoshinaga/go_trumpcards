package domain

import "math/rand"

// memoryCardEntry 記憶したカード1枚分の情報 (ボード上の位置と値)
type memoryCardEntry struct {
	position int // ボード上の位置 0-51
	rank     int // カードの値 1-13
	turnSeen int // 記憶したターン番号
}

// MemoryPlayer 神経衰弱プレイヤークラス
type MemoryPlayer struct {
	*GamePlayer
	pairCount    int               // 獲得したペア数
	pairs        [][2]*Card        // 獲得したペア
	cardMemories []memoryCardEntry // 記憶したカードのリスト
}

// NewMemoryPlayer コンストラクタ
func NewMemoryPlayer(isHuman bool) *MemoryPlayer {
	return &MemoryPlayer{
		GamePlayer:   NewGamePlayer(isHuman),
		pairCount:    0,
		pairs:        nil,
		cardMemories: nil,
	}
}

// GetPairCount 獲得ペア数を取得
func (p *MemoryPlayer) GetPairCount() int { return p.pairCount }

// SetPairCount 獲得ペア数を設定
func (p *MemoryPlayer) SetPairCount(n int) { p.pairCount = n }

// GetPairs 獲得したペア一覧を取得
func (p *MemoryPlayer) GetPairs() [][2]*Card { return p.pairs }

// AddPair ペアを追加
func (p *MemoryPlayer) AddPair(c1, c2 *Card) {
	p.pairs = append(p.pairs, [2]*Card{c1, c2})
	p.pairCount++
}

// ResetMemory カード記憶をリセットする
func (p *MemoryPlayer) ResetMemory() {
	p.cardMemories = nil
}

// ResetGame ゲームリセット（ペア・記憶・手札をクリア）
func (p *MemoryPlayer) ResetGame() {
	p.pairCount = 0
	p.pairs = nil
	p.cardMemories = nil
	p.Reset()
}

// RecordRevealedCard 公開されたカードを記憶する (retentionChance の確率で記録)
// 同じpositionの記憶が既にあれば上書きしない
func (p *MemoryPlayer) RecordRevealedCard(position int, rank int, retentionChance float64, turnNumber int) {
	for _, entry := range p.cardMemories {
		if entry.position == position {
			return // already remembered
		}
	}
	if rand.Float64() < retentionChance {
		p.cardMemories = append(p.cardMemories, memoryCardEntry{position: position, rank: rank, turnSeen: turnNumber})
	}
}

// DecayMemories 古い記憶を確率的に忘却する
func (p *MemoryPlayer) DecayMemories(currentTurn int, decayRate float64) {
	kept := p.cardMemories[:0]
	for _, entry := range p.cardMemories {
		age := currentTurn - entry.turnSeen
		forgetProb := decayRate * float64(age)
		if forgetProb >= 1.0 || rand.Float64() < forgetProb {
			continue
		}
		kept = append(kept, entry)
	}
	p.cardMemories = kept
}

// FindKnownMatch 指定したrankの既知ペア位置を探す
// 同じrankの位置を2つ以上知っていれば最初の2つを返す
func (p *MemoryPlayer) FindKnownMatch(rank int) (int, int, bool) {
	positions := []int{}
	for _, entry := range p.cardMemories {
		if entry.rank == rank {
			positions = append(positions, entry.position)
		}
	}
	if len(positions) >= 2 {
		return positions[0], positions[1], true
	}
	return -1, -1, false
}

// FindAnyKnownPair 記憶の中からペアになる位置の組を探す
func (p *MemoryPlayer) FindAnyKnownPair() (int, int, bool) {
	rankToPositions := map[int][]int{}
	for _, entry := range p.cardMemories {
		rankToPositions[entry.rank] = append(rankToPositions[entry.rank], entry.position)
	}
	for _, positions := range rankToPositions {
		if len(positions) >= 2 {
			return positions[0], positions[1], true
		}
	}
	return -1, -1, false
}

// RemoveMemoryAt 指定位置の記憶を削除する（カードが取られた場合）
func (p *MemoryPlayer) RemoveMemoryAt(position int) {
	kept := p.cardMemories[:0]
	for _, entry := range p.cardMemories {
		if entry.position != position {
			kept = append(kept, entry)
		}
	}
	p.cardMemories = kept
}

// GetMemoryCount 記憶しているカード枚数を返す
func (p *MemoryPlayer) GetMemoryCount() int {
	return len(p.cardMemories)
}
