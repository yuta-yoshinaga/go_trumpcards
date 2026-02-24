package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DaifugoPresenter 大富豪プレゼンターインタフェース
type DaifugoPresenter interface {
	Output(dg interfaces.DaifugoGame, lastErr error) string
}
