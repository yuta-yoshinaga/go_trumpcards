package domain

// memoryManager カード記憶の管理を共通化する埋め込み構造体
type memoryManager[T MemoryEntry] struct {
	cardMemories []T
}

// AddMemory 記憶を追加する
func (m *memoryManager[T]) AddMemory(entry T) {
	m.cardMemories = append(m.cardMemories, entry)
}

// ResetMemory カード記憶をリセットする
func (m *memoryManager[T]) ResetMemory() {
	m.cardMemories = nil
}

// DecayMemories 古い記憶を確率的に忘却する
func (m *memoryManager[T]) DecayMemories(currentTurn int, decayRate float64) {
	m.cardMemories = DecayMemories(m.cardMemories, currentTurn, decayRate)
}
