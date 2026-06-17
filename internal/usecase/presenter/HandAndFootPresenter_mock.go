//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHandAndFootPresenter ハンドアンドフットプレゼンターモック
type MockHandAndFootPresenter = MockGamePresenter[interfaces.HandAndFootGame]
