package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BaseGame 全ゲーム共通のインタフェース
type BaseGame interface {
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
}
