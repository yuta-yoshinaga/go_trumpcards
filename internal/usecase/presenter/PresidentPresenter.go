package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PresidentPresenter プレジデントプレゼンターインタフェース
type PresidentPresenter = GamePresenter[interfaces.PresidentGame]
