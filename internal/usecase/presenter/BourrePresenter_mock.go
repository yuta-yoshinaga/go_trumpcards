//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBourrePresenter ブーレプレゼンターモック
type MockBourrePresenter = MockGamePresenter[interfaces.BourreGame]
