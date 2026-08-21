//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// StalactitesGame フリーセルゲームインタフェース
type StalactitesGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションにカードを移動する
	MoveTableauToFoundation(col int) error
	// MoveTableauToStalactites タブローからフリーセルにカードを移動する
	MoveTableauToStalactites(col, cell int) error
	// MoveStalactitesToTableau フリーセルからタブローにカードを移動する
	MoveStalactitesToTableau(cell, col int) error
	// MoveStalactitesToFoundation フリーセルからファンデーションにカードを移動する
	MoveStalactitesToFoundation(cell int) error
	// GetHint ヒントを取得する
	GetHint() *domain.StalactitesHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.StalactitesPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.StalactitesTableauCnt][]*domain.Card
	// GetMaxMovableCards いま一度に動かせる最大枚数を取得する
	// GetBaseRank はファンデーションの開始ランク（配りごとに変わる）。
	GetBaseRank() int
	GetMaxMovableCards() int
	// GetMaxMovableCardsToEmptyColumn 空き列へ動かすときの上限を取得する
	GetMaxMovableCardsToEmptyColumn() int
	// GetCells フリーセルを取得する
	GetCells() [domain.StalactitesCellCnt]*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.StalactitesFoundationCnt][]*domain.Card
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
