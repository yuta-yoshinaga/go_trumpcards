//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ZhengPresenter 争上游プレゼンターインタフェース
type ZhengPresenter = GamePresenter[interfaces.ZhengGame]
