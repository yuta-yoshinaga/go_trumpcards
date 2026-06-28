package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PrsiPresenter プルシープレゼンターインタフェース
type PrsiPresenter = GamePresenter[interfaces.PrsiGame]
