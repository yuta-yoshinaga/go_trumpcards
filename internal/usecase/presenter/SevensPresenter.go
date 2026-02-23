package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SevensPresenter 7並べプレゼンターインタフェース
type SevensPresenter interface {
	Output(s *domain.Sevens, lastErr error) string
}
