package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OldMaidPresenter ババ抜きプレゼンターインタフェース
type OldMaidPresenter interface {
	Output(om *domain.OldMaid, lastErr error) string
}
