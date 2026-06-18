//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMaoPresenter マオプレゼンターモック
type MockMaoPresenter = MockGamePresenter[interfaces.MaoGame]
