//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BanLuckPresenter バンラックプレゼンターインタフェース
type BanLuckPresenter interface {
	GamePresenter[interfaces.BanLuckGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.BanLuckGame) string
}
