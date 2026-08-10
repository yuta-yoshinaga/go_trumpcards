//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AgnesGame アグネス・ソレルゲームインタフェース
type AgnesGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// DealStock ストックから各列に1枚ずつ配る
	DealStock() error
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col int) error
	// GiveUp ギブアップ
	GiveUp()
	// GetHint ヒント取得
	GetHint() *domain.AgnesHint
	// Undo アンドゥ
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// UndoN n回アンドゥ
	UndoN(n int) error

	// GetPhase フェーズ取得
	GetPhase() domain.AgnesPhase
	// GetMoveCount 移動回数
	GetMoveCount() int
	// GetStockCount ストック残枚数
	GetStockCount() int
	// IsStalemate 合法手が無い状態か
	IsStalemate() bool
	// GetTableau タブロー取得
	GetTableau() [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	// GetFoundation ファンデーション取得
	GetFoundation() [domain.AgnesFoundationCnt][]*domain.Card
	// GetBaseRank ベースランク取得
	GetBaseRank() int
}
