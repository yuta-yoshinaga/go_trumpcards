package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HighCardFlushPresenter ハイカードフラッシュプレゼンターインタフェース
type HighCardFlushPresenter = GamePresenter[interfaces.HighCardFlushGame]
