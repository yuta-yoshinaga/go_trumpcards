//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFiftyOnePresenter フィフティワンプレゼンターモック
type MockFiftyOnePresenter = MockGamePresenter[interfaces.FiftyOneGame]
