//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCuarentaPresenter クアレンタプレゼンターモック。
type MockCuarentaPresenter = MockGamePresenter[interfaces.CuarentaGame]
