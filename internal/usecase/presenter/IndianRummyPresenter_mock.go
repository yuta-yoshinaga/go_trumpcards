//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIndianRummyPresenter インドラミープレゼンターモック
type MockIndianRummyPresenter = MockGamePresenter[interfaces.IndianRummyGame]
