package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HoldemPresenter テキサスホールデムプレゼンターインタフェース
type HoldemPresenter interface {
	Output(h interfaces.HoldemGame, lastErr error) string
	ActionLogOutput(h interfaces.HoldemGame) string
}
