//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTexasHoldemBonusPresenter テキサスホールデムボーナスポーカープレゼンターモック
type MockTexasHoldemBonusPresenter = MockGamePresenter[interfaces.TexasHoldemBonusGame]
