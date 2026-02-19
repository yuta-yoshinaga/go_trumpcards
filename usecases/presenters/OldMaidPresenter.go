package presenters

import "github.com/yuta-yoshinaga/go_trumpcards/entities"

// OldMaidPresenter ババ抜きプレゼンターインタフェース
type OldMaidPresenter interface {
	Output(om *entities.OldMaid) string
}
