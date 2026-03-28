//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCribbagePresenter クリベッジプレゼンターモック
type MockCribbagePresenter = MockGamePresenter[interfaces.CribbageGame]
