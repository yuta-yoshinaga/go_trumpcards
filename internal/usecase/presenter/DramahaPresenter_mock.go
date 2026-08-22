//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDramahaPresenter ドラマハホールデムプレゼンターモック
type MockDramahaPresenter = MockGamePresenter[interfaces.DramahaGame]
