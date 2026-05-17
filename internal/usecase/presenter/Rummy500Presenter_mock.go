//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRummy500Presenter Rummy 500プレゼンターモック
type MockRummy500Presenter = MockGamePresenter[interfaces.Rummy500Game]
