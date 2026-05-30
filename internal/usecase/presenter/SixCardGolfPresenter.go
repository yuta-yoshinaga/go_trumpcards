package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SixCardGolfPresenter シックスカードゴルフプレゼンターインタフェース
type SixCardGolfPresenter = GamePresenter[interfaces.SixCardGolfGame]
