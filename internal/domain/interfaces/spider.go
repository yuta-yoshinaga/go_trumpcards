package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SpiderGame スパイダーソリティアゲームインタフェース
type SpiderGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 指定設定でゲームを初期化する
	ResetWithConfig(cfg domain.SpiderConfig)
	// Deal ストックからタブローに配る
	Deal() error
	// MoveTableauToTableau タブロー間でカードを移動する
	MoveTableauToTableau(fromCol, cardIndex, toCol int) error
	// GiveUp ギブアップする
	GiveUp()
	// GetHint ヒントを取得する
	GetHint() *domain.SpiderHint
	// AutoComplete 自動完了を実行する
	AutoComplete() error
	// Undo 操作を元に戻す
	Undo() error

	// CanUndo 元に戻す操作が可能かを返す
	CanUndo() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SpiderPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetTableau タブローを取得する
	GetTableau() [domain.SpiderTableauCnt][]*domain.SpiderTableauCard
	// GetCompletedSuits 完成スート数を取得する
	GetCompletedSuits() int
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// GetScore 現在のスコアを取得する
	GetScore() int
	// GetDifficulty 難易度を取得する
	GetDifficulty() domain.SpiderDifficulty
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
}
