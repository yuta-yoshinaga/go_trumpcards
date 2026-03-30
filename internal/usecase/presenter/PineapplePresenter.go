package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PineapplePresenter パイナップルポーカープレゼンターインタフェース
type PineapplePresenter = GamePresenter[interfaces.PineappleGame]
