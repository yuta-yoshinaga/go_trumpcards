//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGoFishPresenter Go Fishプレゼンターモック
type MockGoFishPresenter = MockGamePresenter[interfaces.GoFishGame]
