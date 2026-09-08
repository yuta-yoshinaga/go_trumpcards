//go:build !js || !wasm || extra5
// +build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTongitsPresenter Tongitsプレゼンターモック
type MockTongitsPresenter = MockGamePresenter[interfaces.TongitsGame]
