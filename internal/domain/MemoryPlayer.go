package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// memoryCardEntry 記憶したカード1枚分の情報 (ボード上の位置と値)
type memoryCardEntry struct {
	position int // ボード上の位置 0-51
	rank     int // カードの値 1-13
	turnSeen int // 記憶したターン番号
}

// GetTurnSeen MemoryEntryインターフェース実装
func (e memoryCardEntry) GetTurnSeen() int { return e.turnSeen }

// MemoryPlayer 神経衰弱プレイヤークラス
type MemoryPlayer struct {
	*GamePlayer
	pairCount int        // 獲得したペア数
	pairs     [][2]*Card // 獲得したペア
	memoryManager[memoryCardEntry]
}

// NewMemoryPlayer コンストラクタ
func NewMemoryPlayer(isHuman bool) *MemoryPlayer {
	return &MemoryPlayer{
		GamePlayer: NewGamePlayer(isHuman),
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

// ResetGame ゲームリセット（ペア・記憶・手札をクリア）
func (p *MemoryPlayer) ResetGame() {
	p.pairCount = 0
	p.pairs = nil
	p.ResetMemory()
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
		p.AddMemory(memoryCardEntry{position: position, rank: rank, turnSeen: turnNumber})
	}
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

// memoryPlayerJSON is the JSON wire format for MemoryPlayer.
type memoryPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	PairCount  int         `json:"pc"`
	Pairs      [][2]*Card  `json:"pa"`
	// CPU memory (cardMemories) is intentionally omitted;
	// CPU will re-learn after restore.
}

// MarshalJSON implements json.Marshaler.
func (p *MemoryPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(memoryPlayerJSON{
		GamePlayer: p.GamePlayer,
		PairCount:  p.pairCount,
		Pairs:      p.pairs,
	})
}

// memoryPlayerMaxSliceLen caps slice sizes during deserialisation.
const memoryPlayerMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (p *MemoryPlayer) UnmarshalJSON(data []byte) error {
	var j memoryPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Pairs) > memoryPlayerMaxSliceLen {
		return fmt.Errorf("memoryPlayer: input array exceeds maximum allowed size")
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.pairCount = j.PairCount
	p.pairs = j.Pairs
	if p.pairs == nil {
		p.pairs = make([][2]*Card, 0)
	}
	// CPU memory is reset on restore; CPU will re-learn.
	p.ResetMemory()
	return nil
}
