//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SixBidSoloPresenter シックスビッド・ソロ (Six-Bid Solo) プレゼンターインタフェース
type SixBidSoloPresenter interface {
	GamePresenter[interfaces.SixBidSoloGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SixBidSoloGame) string
}
