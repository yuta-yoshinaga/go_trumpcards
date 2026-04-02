package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GoFishPresenter Go Fishプレゼンターインタフェース
type GoFishPresenter = GamePresenter[interfaces.GoFishGame]
