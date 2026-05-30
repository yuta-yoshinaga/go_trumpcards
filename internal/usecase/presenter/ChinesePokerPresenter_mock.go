//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockChinesePokerPresenter チャイニーズポーカープレゼンターモック
type MockChinesePokerPresenter = MockGamePresenter[interfaces.ChinesePokerGame]
