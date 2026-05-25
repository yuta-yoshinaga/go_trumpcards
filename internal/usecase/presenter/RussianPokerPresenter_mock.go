//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRussianPokerPresenter ロシアンポーカープレゼンターモック
type MockRussianPokerPresenter = MockGamePresenter[interfaces.RussianPokerGame]
