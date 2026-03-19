package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OmahaPresenter オマハホールデムプレゼンターインタフェース
type OmahaPresenter = GamePresenter[interfaces.OmahaGame]
