//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKillePresenter キッレ (Kille) プレゼンターモック
type MockKillePresenter = MockGamePresenter[interfaces.KilleGame]
