//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockConquianPresenter コンキャンプレゼンターモック
type MockConquianPresenter = MockGamePresenter[interfaces.ConquianGame]
