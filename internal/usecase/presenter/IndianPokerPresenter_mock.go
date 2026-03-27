//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIndianPokerPresenter インディアンポーカープレゼンターモック
type MockIndianPokerPresenter = MockGamePresenter[interfaces.IndianPokerGame]
