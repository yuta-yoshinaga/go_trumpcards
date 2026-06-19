//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OpenFaceChinesePresenter オープンフェイス・チャイニーズポーカー (OFC) のプレゼンターインタフェース
type OpenFaceChinesePresenter interface {
	GamePresenter[interfaces.OpenFaceChineseGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.OpenFaceChineseGame) string
}
