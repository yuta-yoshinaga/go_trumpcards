//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RussianBankPresenter ロシアンバンク (クラペット) のプレゼンターインタフェース。
type RussianBankPresenter interface {
	GamePresenter[interfaces.RussianBankGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.RussianBankGame) string
}
