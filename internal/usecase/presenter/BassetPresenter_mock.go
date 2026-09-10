//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBassetPresenter はバセットプレゼンターのモック。
type MockBassetPresenter = MockGamePresenter[interfaces.BassetGame]
