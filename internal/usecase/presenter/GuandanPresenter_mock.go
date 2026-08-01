//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGuandanPresenter 掼蛋 (Guandan) プレゼンターモック
type MockGuandanPresenter = MockGamePresenter[interfaces.GuandanGame]
