//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTongitsPresenter Tongitsプレゼンターモック
type MockTongitsPresenter = MockGamePresenter[interfaces.TongitsGame]
