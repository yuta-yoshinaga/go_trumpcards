//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTichuPresenter ティチュープレゼンターモック
type MockTichuPresenter = MockGamePresenter[interfaces.TichuGame]
