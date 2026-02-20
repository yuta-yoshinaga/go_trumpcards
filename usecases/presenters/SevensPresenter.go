package presenters

import "github.com/yuta-yoshinaga/go_trumpcards/entities"

// SevensPresenter 7並べプレゼンターインタフェース
type SevensPresenter interface {
	Output(s *entities.Sevens) string
}
