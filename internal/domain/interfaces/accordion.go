//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AccordionGame アコーディオンゲームインタフェース
type AccordionGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Move fromIdx のパイルを toIdx のパイルに重ねる
	Move(fromIdx, toIdx int) error
	// GiveUp ギブアップする
	GiveUp()
	// AutoComplete ヒントが示す手を尽きるまで繰り返す
	AutoComplete() error
	// GetHint ヒントを取得する
	GetHint() *domain.AccordionHint
	// Undo 操作を元に戻す
	Undo() error
	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
	UndoToEscape() int
	// UndoN n回連続でアンドゥを実行する
	UndoN(n int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AccordionPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetPiles パイル一覧を取得する
	GetPiles() [][]*domain.Card
	// GetPileCount 残りパイル数を取得する
	GetPileCount() int
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
