//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DoudizhuPresenter 斗地主プレゼンターインタフェース
type DoudizhuPresenter = GamePresenter[interfaces.DoudizhuGame]
