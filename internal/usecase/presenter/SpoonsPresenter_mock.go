//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpoonsPresenter はスプーンのプレゼンターモック。
type MockSpoonsPresenter = MockGamePresenter[interfaces.SpoonsGame]
