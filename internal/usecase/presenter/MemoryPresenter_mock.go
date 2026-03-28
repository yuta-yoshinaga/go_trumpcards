//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMemoryPresenter 神経衰弱プレゼンターモック
type MockMemoryPresenter = MockGamePresenter[interfaces.MemoryGame]
