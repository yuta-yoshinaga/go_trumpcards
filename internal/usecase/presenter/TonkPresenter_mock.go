//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTonkPresenter Tonkプレゼンターモック
type MockTonkPresenter = MockGamePresenter[interfaces.TonkGame]
