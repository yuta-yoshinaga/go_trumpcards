//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OichoKabuPresenter おいちょかぶプレゼンターインタフェース
type OichoKabuPresenter = GamePresenter[interfaces.OichoKabuGame]
