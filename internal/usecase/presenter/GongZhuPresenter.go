package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GongZhuPresenter 拱猪プレゼンターインタフェース
type GongZhuPresenter interface {
	GamePresenter[interfaces.GongZhuGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GongZhuGame) string
}
