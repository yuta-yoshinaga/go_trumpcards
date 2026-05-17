//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFourCardPokerPresenter is the mock presenter used in usecase-layer tests.
type MockFourCardPokerPresenter = MockGamePresenter[interfaces.FourCardPokerGame]
