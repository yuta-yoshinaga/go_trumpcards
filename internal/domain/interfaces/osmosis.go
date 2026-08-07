//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OsmosisGame オズモシス（浸透）ソリティアゲームインタフェース
type OsmosisGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation(fIdx int) error
	// MoveReserveToFoundation リザーブからファンデーションに移動
	MoveReserveToFoundation(rIdx, fIdx int) error
	// GiveUp ギブアップ
	GiveUp()
	// GetHint ヒント取得
	GetHint() *domain.OsmosisHint
	// AutoComplete 自動完了
	AutoComplete() error
	// Undo アンドゥ
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// UndoN n回アンドゥ
	UndoN(n int) error

	// GetPhase フェーズ取得
	GetPhase() domain.OsmosisPhase
	// GetMoveCount 移動回数
	GetMoveCount() int
	// GetStockCount ストック残枚数
	GetStockCount() int
	// GetWaste ウェイスト取得
	GetWaste() []*domain.Card
	// GetReserve リザーブ取得（4列）
	GetReserve() [domain.OsmosisReserveCnt][]*domain.Card
	// GetFoundation ファンデーション取得
	GetFoundation() [domain.OsmosisFoundationCnt][]*domain.Card
	// GetBaseRank ベースランク取得
	GetBaseRank() int
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
