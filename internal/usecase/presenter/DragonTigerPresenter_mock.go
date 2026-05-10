//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDragonTigerPresenter ドラゴンタイガープレゼンターモック
type MockDragonTigerPresenter = MockGamePresenter[interfaces.DragonTigerGame]
