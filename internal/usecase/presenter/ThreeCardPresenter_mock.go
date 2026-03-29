//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThreeCardPresenter スリーカードポーカープレゼンターモック
type MockThreeCardPresenter = MockGamePresenter[interfaces.ThreeCardGame]
