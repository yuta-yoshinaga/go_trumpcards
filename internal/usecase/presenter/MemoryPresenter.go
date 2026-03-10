package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MemoryPresenter 神経衰弱プレゼンターインタフェース
type MemoryPresenter interface {
	Output(m interfaces.MemoryGame, lastErr error) string
	ActionLogOutput(m interfaces.MemoryGame) string
}
