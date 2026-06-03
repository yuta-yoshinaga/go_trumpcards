//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBurracoPresenter ブラーコプレゼンターモック
type MockBurracoPresenter = MockGamePresenter[interfaces.BurracoGame]
