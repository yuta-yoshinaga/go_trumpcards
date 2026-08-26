//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SergeantMajorPresenter サージェントメジャープレゼンターインタフェース
type SergeantMajorPresenter interface {
	GamePresenter[interfaces.SergeantMajorGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SergeantMajorGame) string
}
