//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBlackJackPresenter ブラックジャックプレゼンターモック
type MockBlackJackPresenter = MockGamePresenter[interfaces.BlackJackGame]
