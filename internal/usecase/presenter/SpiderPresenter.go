package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpiderPresenter スパイダーソリティアプレゼンターインタフェース
type SpiderPresenter interface {
	GamePresenter[interfaces.SpiderGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SpiderGame) string
}
