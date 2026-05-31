//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockScopaPresenter スコパプレゼンターモック。
type MockScopaPresenter = MockGamePresenter[interfaces.ScopaGame]
