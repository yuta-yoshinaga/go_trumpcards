//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBadugiPresenter is the testify/mock presenter for Badugi.
type MockBadugiPresenter = MockGamePresenter[interfaces.BadugiGame]
