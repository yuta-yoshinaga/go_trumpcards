package presenters

import "github.com/yuta-yoshinaga/go_trumpcards/entities"

// DaifugoPresenter 大富豪プレゼンターインタフェース
type DaifugoPresenter interface {
	Output(dg *entities.Daifugo) string
}
