//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPishtiPresenter は Pişti プレゼンターモック。
type MockPishtiPresenter = MockGamePresenter[interfaces.PishtiGame]
