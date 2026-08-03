//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// WindmillGame ウィンドミル ゲームインタフェース
type WindmillGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() error
	// MoveSailToCenter 帆の札を中央基礎札へ送る
	MoveSailToCenter(sailIdx int) error
	// MoveSailToCorner 帆の札を四隅の基礎札へ送る
	MoveSailToCorner(sailIdx, cornerIdx int) error
	// MoveWasteToCenter 捨て札の一番上を中央基礎札へ送る
	MoveWasteToCenter() error
	// MoveWasteToCorner 捨て札の一番上を四隅の基礎札へ送る
	MoveWasteToCorner(cornerIdx int) error
	// MoveCornerToCenter 四隅の一番上を中央基礎札へ引き戻す
	MoveCornerToCenter(cornerIdx int) error
	// GetHint ヒントを取得する
	GetHint() *domain.WindmillHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.WindmillPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetSails 十字（帆）の 8 枠を取得する
	GetSails() [domain.WindmillSailCnt]*domain.Card
	// GetCenter 中央基礎札を取得する
	GetCenter() []*domain.Card
	// GetCorners 四隅の基礎札を取得する
	GetCorners() [domain.WindmillCornerCnt][]*domain.Card
	// IsTransferBlocked 四隅→中央の引き戻しが今は禁じられているか
	IsTransferBlocked() bool
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
