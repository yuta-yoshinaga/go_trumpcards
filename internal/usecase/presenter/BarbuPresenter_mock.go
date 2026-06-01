//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBarbuPresenter はバルブプレゼンターモック。
type MockBarbuPresenter = MockGamePresenter[interfaces.BarbuGame]
