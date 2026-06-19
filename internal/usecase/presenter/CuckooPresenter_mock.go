//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCuckooPresenter Cuckoo プレゼンターモック
type MockCuckooPresenter = MockGamePresenter[interfaces.CuckooGame]
