//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockChinchonPresenter チンチョンプレゼンターモック
type MockChinchonPresenter = MockGamePresenter[interfaces.ChinchonGame]
