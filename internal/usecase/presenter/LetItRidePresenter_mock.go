//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLetItRidePresenter レット・イット・ライドプレゼンターモック
type MockLetItRidePresenter = MockGamePresenter[interfaces.LetItRideGame]
