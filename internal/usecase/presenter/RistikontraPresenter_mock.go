//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRistikontraPresenter はリスティコントラ プレゼンターモック。
type MockRistikontraPresenter = MockGamePresenter[interfaces.RistikontraGame]
