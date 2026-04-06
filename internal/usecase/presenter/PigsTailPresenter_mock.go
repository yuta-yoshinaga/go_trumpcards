//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPigsTailPresenter ぶたのしっぽプレゼンターモック
type MockPigsTailPresenter = MockGamePresenter[interfaces.PigsTailGame]
