package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DaifugoPresenter 大富豪プレゼンターインタフェース
type DaifugoPresenter interface {
	Output(dg *domain.Daifugo) string
}
