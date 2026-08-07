//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CaribbeanStudPresenter カリビアンスタッドポーカープレゼンターインタフェース
type CaribbeanStudPresenter interface {
	GamePresenter[interfaces.CaribbeanStudGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cs interfaces.CaribbeanStudGame) string
}
