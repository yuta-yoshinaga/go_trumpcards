//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSambaPresenter サンバプレゼンターモック
type MockSambaPresenter = MockGamePresenter[interfaces.SambaGame]
