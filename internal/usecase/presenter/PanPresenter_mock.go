//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPanPresenter パングインゲプレゼンターモック
type MockPanPresenter = MockGamePresenter[interfaces.PanGame]
