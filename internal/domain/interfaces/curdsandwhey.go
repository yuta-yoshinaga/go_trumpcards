//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CurdsAndWheyGame カーズ・アンド・ホエイ (スパイダー系ソリティア) のゲームインタフェース。
type CurdsAndWheyGame interface {
	BaseGame
	// GetGameEndFlag プレイ中でなければ true。
	GetGameEndFlag() bool
	// Reset ゲームを初期化する。
	Reset()
	// MoveSequence 列 fromCol の cardIndex 以降を列 toCol へ移す。
	MoveSequence(fromCol, cardIndex, toCol int) error
	// GiveUp 投了する。
	GiveUp()
	// Undo 直近の手を取り消す。
	Undo() error
	// CanUndo Undo 可能か。
	CanUndo() bool
	// UndoN n 回 Undo する。
	UndoN(n int) error
	// GetHint ヒントを取得する。
	GetHint() *domain.CurdsAndWheyHint

	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.CurdsAndWheyPhase
	// GetMoveCount 累計手数を取得する。
	GetMoveCount() int
	// GetCompletedSuits 完成スート数を取得する。
	GetCompletedSuits() int
	// GetColumns タブロー列を取得する。
	GetColumns() [domain.CurdsAndWheyColCnt][]*domain.Card
}
