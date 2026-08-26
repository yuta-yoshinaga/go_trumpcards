//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GermanSoloPresenter ジャーマン・ソロ (GermanSolo) のプレゼンターインタフェース
type GermanSoloPresenter interface {
	GamePresenter[interfaces.GermanSoloGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GermanSoloGame) string
}
