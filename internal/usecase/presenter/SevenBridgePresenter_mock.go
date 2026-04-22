//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSevenBridgePresenter セブンブリッジプレゼンターモック
type MockSevenBridgePresenter = MockGamePresenter[interfaces.SevenBridgeGame]
