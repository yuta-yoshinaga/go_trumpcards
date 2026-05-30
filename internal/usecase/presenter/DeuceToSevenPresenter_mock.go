//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDeuceToSevenPresenter is the testify/mock presenter for 2-7 Triple Draw.
type MockDeuceToSevenPresenter = MockGamePresenter[interfaces.DeuceToSevenGame]
