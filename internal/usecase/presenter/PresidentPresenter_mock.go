//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPresidentPresenter プレジデントプレゼンターモック
type MockPresidentPresenter = MockGamePresenter[interfaces.PresidentGame]
