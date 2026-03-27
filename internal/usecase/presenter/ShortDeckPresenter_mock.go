//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockShortDeckPresenter ショートデックホールデムプレゼンターモック
type MockShortDeckPresenter = MockGamePresenter[interfaces.ShortDeckGame]
