//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GapsGame はGaps（Montana）ゲームのインタフェース。
type GapsGame interface {
	BaseGame
	// GetGameEndFlag はプレイ中以外かどうかを返す。
	GetGameEndFlag() bool
	// Reset はゲームを初期化する。
	Reset()
	// Move はカードを (fromRow,fromCol) から (toRow,toCol) の隙間へ移動する。
	Move(fromRow, fromCol, toRow, toCol int) error
	// Redeal は手詰まり時にカードを再配りする。
	Redeal() error
	// Undo は直前の操作を元に戻す。
	Undo() error
	// UndoN はn回連続でアンドゥする。
	UndoN(n int) error
	// CanUndo はアンドゥ可能かどうかを返す。
	CanUndo() bool
	// UndoToEscape は手詰まりから抜けるために必要なアンドゥ回数を返す。
	UndoToEscape() int
	// GiveUp はギブアップする。
	GiveUp()
	// GetHint はヒントを返す。プレイ中で手があればポインタ、なければnil。
	GetHint() *domain.GapsHint
	// GetPhase はフェーズを返す。
	GetPhase() domain.GapsPhase
	// GetMoveCount は移動回数を返す。
	GetMoveCount() int
	// GetGrid は盤面を返す。
	GetGrid() [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell
	// GetGapNeed そのギャップに何が入るかを取得する (nil=対象外/未確定)
	GetGapNeed(row, col int) *domain.GapsGapNeed
	// GetLockedPrefixLengths は各行のロック済み接頭辞の長さ（再配布で保持される先頭カード数）を返す。
	GetLockedPrefixLengths() [domain.GapsRowCnt]int
	// GetRedealsUsed は使用済み再配り回数を返す。
	GetRedealsUsed() int
	// GetRedealsRemaining は残りの再配り回数を返す。
	GetRedealsRemaining() int
	// IsStalemate は手詰まりかどうかを返す。
	IsStalemate() bool
	// AllWon は勝利状態かどうかを返す。
	AllWon() bool
}
