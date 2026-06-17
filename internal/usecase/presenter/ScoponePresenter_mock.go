//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockScoponePresenter スコポーネプレゼンターモック。
type MockScoponePresenter = MockGamePresenter[interfaces.ScoponeGame]
