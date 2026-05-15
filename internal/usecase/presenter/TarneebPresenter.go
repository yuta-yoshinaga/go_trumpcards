package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TarneebPresenter Tarneeb プレゼンターインタフェース
type TarneebPresenter interface {
	GamePresenter[interfaces.TarneebGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TarneebGame) string
}
