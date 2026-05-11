//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockUltimateTexasHoldemPresenter アルティメット・テキサスホールデムプレゼンターモック
type MockUltimateTexasHoldemPresenter = MockGamePresenter[interfaces.UltimateTexasHoldemGame]
