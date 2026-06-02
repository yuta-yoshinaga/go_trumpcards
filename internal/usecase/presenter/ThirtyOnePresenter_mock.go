//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThirtyOnePresenter ThirtyOne プレゼンターモック
type MockThirtyOnePresenter = MockGamePresenter[interfaces.ThirtyOneGame]
