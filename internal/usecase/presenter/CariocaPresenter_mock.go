//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCariocaPresenter カリオカプレゼンターモック
type MockCariocaPresenter = MockGamePresenter[interfaces.CariocaGame]
