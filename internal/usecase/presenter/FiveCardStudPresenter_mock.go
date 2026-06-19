//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFiveCardStudPresenter ファイブカードスタッドプレゼンターモック
type MockFiveCardStudPresenter = MockGamePresenter[interfaces.FiveCardStudGame]
