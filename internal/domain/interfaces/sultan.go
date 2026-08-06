//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SultanGame スルタンゲームインタフェース
type SultanGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// Redeal ウェイストを集めて新しいストックを作る（最大2回）
	Redeal() error
	// MoveDivanToFoundation ディヴァンからファンデーションにカードを移動する
	MoveDivanToFoundation(divanIdx int) error
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動する
	MoveWasteToFoundation() error
	// GetHint ヒントを取得する
	GetHint() *domain.SultanHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SultanPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste ウェイストのカード一覧を取得する
	GetWaste() []*domain.Card
	// GetDivan ディヴァンを取得する
	GetDivan() []*domain.Card
	// GetFoundation ファンデーションを取得する
	GetFoundation() [domain.SultanFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが可視かを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりを脱出するのに必要なアンドゥ回数 (不明なら -1)
	UndoToEscape() int
	// GetRedealCount リディール使用回数を返す
	GetRedealCount() int
	// CanRedeal リディール可能かどうかを返す
	CanRedeal() bool
}
