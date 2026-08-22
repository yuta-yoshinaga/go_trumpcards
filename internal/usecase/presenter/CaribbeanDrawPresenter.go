//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CaribbeanDrawPresenter カリビアン・ドロー・ポーカープレゼンターインタフェース
type CaribbeanDrawPresenter interface {
	GamePresenter[interfaces.CaribbeanDrawGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cs interfaces.CaribbeanDrawGame) string
}
