//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KlaberjassPresenter クラバーヤス (Klaberjass) プレゼンターインタフェース
type KlaberjassPresenter = GamePresenter[interfaces.KlaberjassGame]
