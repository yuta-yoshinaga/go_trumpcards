//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOasisPokerPresenter オアシスポーカープレゼンターモック
type MockOasisPokerPresenter = MockGamePresenter[interfaces.OasisPokerGame]
