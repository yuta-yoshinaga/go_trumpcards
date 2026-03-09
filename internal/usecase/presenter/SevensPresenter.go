package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SevensPresenter 7並べプレゼンターインタフェース
type SevensPresenter interface {
	Output(s interfaces.SevensGame, lastErr error) string
	ActionLogOutput(s interfaces.SevensGame) string
}
