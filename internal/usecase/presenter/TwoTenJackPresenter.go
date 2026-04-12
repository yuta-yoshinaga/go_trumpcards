package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TwoTenJackPresenter ツーテンジャックプレゼンターインタフェース
type TwoTenJackPresenter interface {
	GamePresenter[interfaces.TwoTenJackGame]
	// HintOutput ヒント情報を出力する
	HintOutput(ttj interfaces.TwoTenJackGame) string
}
