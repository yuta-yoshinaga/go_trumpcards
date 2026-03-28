//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDaifugoPresenter 大富豪プレゼンターモック
type MockDaifugoPresenter = MockGamePresenter[interfaces.DaifugoGame]
