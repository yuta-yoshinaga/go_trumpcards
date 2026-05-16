//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHighCardFlushPresenter ハイカードフラッシュプレゼンターモック
type MockHighCardFlushPresenter = MockGamePresenter[interfaces.HighCardFlushGame]
