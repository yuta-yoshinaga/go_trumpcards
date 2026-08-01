//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockShengJiPresenter 升级 (Sheng Ji) プレゼンターモック
type MockShengJiPresenter = MockGamePresenter[interfaces.ShengJiGame]
