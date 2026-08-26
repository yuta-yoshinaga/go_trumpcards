//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SevenTwentySevenPresenter はセブン・トゥエンティセブン (SevenTwentySeven) のプレゼンターインタフェース。
type SevenTwentySevenPresenter interface {
	GamePresenter[interfaces.SevenTwentySevenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SevenTwentySevenGame) string
}
