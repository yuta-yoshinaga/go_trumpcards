//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBlackJackSwitchPresenter ブラックジャック・スイッチプレゼンターモック
type MockBlackJackSwitchPresenter = MockGamePresenter[interfaces.BlackJackSwitchGame]
