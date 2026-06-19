//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CuarentaPresenter クアレンタプレゼンターインタフェース。
type CuarentaPresenter = GamePresenter[interfaces.CuarentaGame]
