//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSevenCardStudPresenter セブンカードスタッドプレゼンターモック
type MockSevenCardStudPresenter = MockGamePresenter[interfaces.SevenCardStudGame]
