//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CourtPiecePresenter Court Piece プレゼンターインタフェース
type CourtPiecePresenter interface {
	GamePresenter[interfaces.CourtPieceGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.CourtPieceGame) string
}
