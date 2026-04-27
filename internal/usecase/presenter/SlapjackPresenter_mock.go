//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSlapjackPresenter スラップジャックプレゼンターモック
type MockSlapjackPresenter = MockGamePresenter[interfaces.SlapjackGame]
