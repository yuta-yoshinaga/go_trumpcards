package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DoubtPresenter ダウトプレゼンターインタフェース
type DoubtPresenter interface {
	Output(d interfaces.DoubtGame, lastErr error) string
	ActionLogOutput(d interfaces.DoubtGame) string
}
