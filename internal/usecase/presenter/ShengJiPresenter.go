//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ShengJiPresenter 升级 (Sheng Ji) プレゼンターインタフェース
type ShengJiPresenter = GamePresenter[interfaces.ShengJiGame]
