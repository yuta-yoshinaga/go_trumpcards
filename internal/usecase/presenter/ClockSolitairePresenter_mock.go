//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockClockSolitairePresenter クロックソリティアプレゼンターモック
type MockClockSolitairePresenter = MockGamePresenter[interfaces.ClockSolitaireGame]
