//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OsmosisPresenter オズモシスプレゼンターインタフェース
type OsmosisPresenter interface {
	GamePresenter[interfaces.OsmosisGame]
	// HintOutput ヒント情報を出力する
	HintOutput(o interfaces.OsmosisGame) string
}
