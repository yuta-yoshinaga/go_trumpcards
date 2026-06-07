//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockYanivPresenter Yaniv プレゼンターモック
type MockYanivPresenter = MockGamePresenter[interfaces.YanivGame]
