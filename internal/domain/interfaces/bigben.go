//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BigBenGame ビッグ・ベン ゲームインタフェース
type BigBenGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// MoveTableauToFoundation タブローから文字盤へ移動する
	MoveTableauToFoundation(col, fIdx int) error
	// Deal 山札から補充する（各列 3 枚まで、全列 3 枚以上なら 1 巡）
	Deal() error
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// MoveTableauToTableau タブロー間で移動する
	MoveTableauToTableau(fromCol, toCol int) error
	// GetHint ヒントを取得する
	GetHint() *domain.BigBenHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BigBenPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetFoundation 文字盤を取得する
	GetFoundation() [domain.BigBenFoundationCnt][]*domain.Card
	// GetTableau タブローを取得する
	GetTableau() [domain.BigBenTableauCnt][]*domain.BigBenTableauCard
	// IsFoundationComplete 文字盤が目標ランクに達しているか
	IsFoundationComplete(fIdx int) bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
