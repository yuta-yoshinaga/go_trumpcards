package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// NertzGame Nertz / Pounce ゲームインタフェース
type NertzGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 設定を適用してゲームを初期化する
	ResetWithConfig(cfg domain.NertzConfig)
	// NextRound 次ラウンドを開始する
	NextRound()
	// DrawStock ストックからウェイストへカードをめくる
	DrawStock(playerIdx int) error
	// MoveNertzToFoundation ナッツパイル → ファウンデーション
	MoveNertzToFoundation(playerIdx, foundationIdx int) error
	// MoveNertzToTableau ナッツパイル → タブロー
	MoveNertzToTableau(playerIdx, toCol int) error
	// MoveWasteToFoundation ウェイスト → ファウンデーション
	MoveWasteToFoundation(playerIdx, foundationIdx int) error
	// MoveWasteToTableau ウェイスト → タブロー
	MoveWasteToTableau(playerIdx, toCol int) error
	// MoveTableauToFoundation タブロー → ファウンデーション
	MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) error
	// MoveTableauToTableau タブロー → タブロー
	MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) error
	// Tick CPU を 1tick 進めて適用済みアクションを返す
	Tick() []*domain.NertzAction
	// GetHint 人間プレイヤーの推奨手を返す
	GetHint() *domain.NertzHint
	// CanUndo Undo 可能か
	CanUndo() bool
	// Undo 直前の人間アクションを取り消す
	Undo() error
	// GetPhase 現在のフェーズ
	GetPhase() domain.NertzPhase
	// GetRoundNo ラウンド番号 (1始まり)
	GetRoundNo() int
	// GetWinnerIdx ラウンド勝者
	GetWinnerIdx() int
	// GetMatchWinner マッチ勝者
	GetMatchWinner() int
	// GetConfig 設定
	GetConfig() domain.NertzConfig
	// GetPlayers プレイヤースナップショット
	GetPlayers() []*domain.NertzPlayer
	// GetFoundations ファウンデーションスナップショット
	GetFoundations() []*domain.NertzFoundation
	// GetMoveCount 通算手数
	GetMoveCount() int
}
