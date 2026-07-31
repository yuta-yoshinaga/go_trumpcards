//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockVintPresenter ヴィント (Vint) プレゼンターモック
type MockVintPresenter = MockGamePresenter[interfaces.VintGame]
