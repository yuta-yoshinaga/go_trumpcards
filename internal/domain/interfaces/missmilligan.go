//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MissMilliganGame ミス・ミリガン ゲームインタフェース
type MissMilliganGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Deal 山札から各列へ 1 枚ずつ配り足す
	Deal() error
	// MoveTableauToTableau タブロー間で移動する（cardIndex 以降の連番をまとめて）
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// MoveTableauToFoundation タブローから基礎札へ移動する
	MoveTableauToFoundation(col int) error
	// Waive タブローの連番を脇へ持ち上げる
	Waive(col, cardIndex int) error
	// PlaceWaived 保持中の札をタブローへ戻す
	PlaceWaived(toCol int) error
	// MoveWaivedToFoundation 保持中の 1 枚を基礎札へ送る
	MoveWaivedToFoundation() error
	// GetHint ヒントを取得する
	GetHint() *domain.MissMilliganHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MissMilliganPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaived 保持中の札を取得する
	GetWaived() []*domain.Card
	// CanWaive ウェイブが可能か
	CanWaive() bool
	// GetTableau タブローを取得する
	GetTableau() [domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.MissMilliganFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
