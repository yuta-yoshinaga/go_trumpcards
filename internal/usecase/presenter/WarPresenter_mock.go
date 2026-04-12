//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWarPresenter 戦争プレゼンターモック
type MockWarPresenter = MockGamePresenter[interfaces.WarGame]
