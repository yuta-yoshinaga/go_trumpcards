package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OldMaidPresenter ババ抜きプレゼンターインタフェース
type OldMaidPresenter interface {
	Output(om interfaces.OldMaidGame, lastErr error) string
}
