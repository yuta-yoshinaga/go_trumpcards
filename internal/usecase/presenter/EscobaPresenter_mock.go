//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEscobaPresenter エスコバプレゼンターモック。
type MockEscobaPresenter = MockGamePresenter[interfaces.EscobaGame]
