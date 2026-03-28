//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockVideoPokerPresenter ビデオポーカープレゼンターモック
type MockVideoPokerPresenter = MockGamePresenter[interfaces.VideoPokerGame]
