//go:build !js || !wasm || extra11

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MonteBankPresenter モンテバンクプレゼンターインタフェース
type MonteBankPresenter interface {
	GamePresenter[interfaces.MonteBankGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.MonteBankGame) string
}
