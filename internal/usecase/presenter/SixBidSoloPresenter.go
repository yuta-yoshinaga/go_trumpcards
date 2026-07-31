//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SixBidSoloPresenter シックスビッド・ソロ (Six-Bid Solo) プレゼンターインタフェース
type SixBidSoloPresenter = GamePresenter[interfaces.SixBidSoloGame]
