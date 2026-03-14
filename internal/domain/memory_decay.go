package domain

import "math/rand"

// MemoryEntry turnSeen を持つ記憶エントリのインターフェース
type MemoryEntry interface {
	GetTurnSeen() int
}

// DecayMemories 古い記憶を確率的に忘却する共通関数
// 忘却確率 = decayRate * (currentTurn - entry.GetTurnSeen())
func DecayMemories[T MemoryEntry](memories []T, currentTurn int, decayRate float64) []T {
	if len(memories) == 0 || decayRate <= 0 {
		return memories
	}
	kept := memories[:0]
	for _, entry := range memories {
		age := currentTurn - entry.GetTurnSeen()
		forgetProb := decayRate * float64(age)
		if forgetProb >= 1.0 || rand.Float64() < forgetProb {
			continue
		}
		kept = append(kept, entry)
	}
	return kept
}
