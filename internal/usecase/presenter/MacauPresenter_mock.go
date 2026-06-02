//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMacauPresenter マカオプレゼンターモック
type MockMacauPresenter = MockGamePresenter[interfaces.MacauGame]
