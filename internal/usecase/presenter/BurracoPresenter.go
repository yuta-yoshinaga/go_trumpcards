package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BurracoPresenter ブラーコプレゼンタインタフェース
type BurracoPresenter = GamePresenter[interfaces.BurracoGame]
