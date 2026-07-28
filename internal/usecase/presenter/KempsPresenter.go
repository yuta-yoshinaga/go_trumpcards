//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KempsPresenter はケムプスのプレゼンターインタフェース。
type KempsPresenter = GamePresenter[interfaces.KempsGame]
