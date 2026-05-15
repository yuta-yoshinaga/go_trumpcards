//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCasinoHoldemPresenter カジノホールデムプレゼンターモック
type MockCasinoHoldemPresenter = MockGamePresenter[interfaces.CasinoHoldemGame]
