//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPineapplePresenter パイナップルポーカープレゼンターモック
type MockPineapplePresenter = MockGamePresenter[interfaces.PineappleGame]
