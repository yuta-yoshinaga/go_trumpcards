//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRistikontraPresenter は Pişti プレゼンターモック。
type MockRistikontraPresenter = MockGamePresenter[interfaces.RistikontraGame]
