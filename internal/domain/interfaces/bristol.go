//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BristolGame ブリストルソリティアゲームインタフェース
type BristolGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw ストックから3つのファンに1枚ずつ配る
	Draw() error
	// MoveTableauToTableau タブローからタブローに移動
	MoveTableauToTableau(fromCol, toCol int) error
	// MoveTableauToFoundation タブローからファウンデーションに移動
	MoveTableauToFoundation(col int) error
	// MoveFanToTableau ファンからタブローに移動
	MoveFanToTableau(fanIdx, toCol int) error
	// MoveFanToFoundation ファンからファウンデーションに移動
	MoveFanToFoundation(fanIdx int) error
	// GiveUp ギブアップ
	GiveUp()
	// GetHint ヒント取得
	GetHint() *domain.BristolHint
	// AutoComplete 自動完了
	AutoComplete() error
	// Undo アンドゥ
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// UndoN n回アンドゥ
	UndoN(n int) error

	// GetPhase フェーズ取得
	GetPhase() domain.BristolPhase
	// GetMoveCount 移動回数
	GetMoveCount() int
	// GetStockCount ストック残枚数
	GetStockCount() int
	// GetTableau タブロー取得（8列）
	GetTableau() [domain.BristolTableauCnt][]*domain.Card
	// GetFan ファン取得（3つ）
	GetFan() [domain.BristolFanCnt][]*domain.Card
	// GetFoundation ファウンデーション取得（4つ）
	GetFoundation() [domain.BristolFoundationCnt][]*domain.Card
	// BristolLegalTargets 指定した移動元の札を置ける先 (タブロー列/ファウンデーション)
	BristolLegalTargets(fromZone string, fromCol int) ([]int, []int)
}
