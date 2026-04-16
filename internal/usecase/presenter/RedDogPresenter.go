package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RedDogPresenter レッドドッグプレゼンターインタフェース
type RedDogPresenter = GamePresenter[interfaces.RedDogGame]
