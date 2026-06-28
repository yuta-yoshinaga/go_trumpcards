//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPrsiPresenter プルシープレゼンターモック
type MockPrsiPresenter = MockGamePresenter[interfaces.PrsiGame]
