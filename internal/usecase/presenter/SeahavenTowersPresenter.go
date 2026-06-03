//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SeahavenTowersPresenter シーヘイブンタワーズプレゼンターインタフェース
type SeahavenTowersPresenter interface {
	GamePresenter[interfaces.SeahavenTowersGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SeahavenTowersGame) string
}
