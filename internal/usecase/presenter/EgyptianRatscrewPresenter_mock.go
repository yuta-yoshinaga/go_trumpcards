//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEgyptianRatscrewPresenter エジプシャン・ラットスクリュープレゼンターモック
type MockEgyptianRatscrewPresenter = MockGamePresenter[interfaces.EgyptianRatscrewGame]
