//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// StreetsAndAlleysPresenter Streets and Alleys プレゼンターインタフェース
type StreetsAndAlleysPresenter interface {
	GamePresenter[interfaces.StreetsAndAlleysGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.StreetsAndAlleysGame) string
}
