//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LiteraturePresenter リテラチャー (Literature) プレゼンターインタフェース
type LiteraturePresenter = GamePresenter[interfaces.LiteratureGame]
