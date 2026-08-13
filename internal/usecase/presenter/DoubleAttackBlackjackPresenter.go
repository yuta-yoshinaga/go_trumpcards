//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DoubleAttackBlackjackPresenter 追加ベット・ブラックジャックプレゼンターインタフェース
type DoubleAttackBlackjackPresenter interface {
	GamePresenter[interfaces.DoubleAttackBlackjackGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.DoubleAttackBlackjackGame) string
}
