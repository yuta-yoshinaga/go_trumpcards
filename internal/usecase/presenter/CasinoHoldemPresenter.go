//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CasinoHoldemPresenter カジノホールデムプレゼンターインタフェース
type CasinoHoldemPresenter interface {
	GamePresenter[interfaces.CasinoHoldemGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CasinoHoldemGame) string
}
