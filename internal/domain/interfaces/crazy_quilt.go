//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CrazyQuiltGame クレイジーキルト ゲームインタフェース
type CrazyQuiltGame interface {
	SolitaireGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() error
	// MoveQuiltToFoundation キルトの札を基礎札へ送る
	MoveQuiltToFoundation(idx int) error
	// MoveQuiltToWaste キルトの札を捨て札の上へ置く
	MoveQuiltToWaste(idx int) error
	// MoveWasteToFoundation 捨て札から基礎札へ移動する
	MoveWasteToFoundation() error
	// IsAvailable そのマスの札が取れるか（短辺が露出しているか）
	IsAvailable(idx int) bool
	// GetRedealsLeft 残りの組み直し回数を取得する
	GetRedealsLeft() int
	// IsAscendingFoundation その基礎札が A からの昇順かを返す
	IsAscendingFoundation(fIdx int) bool
	// GetHint ヒントを取得する
	GetHint() *domain.CrazyQuiltHint
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CrazyQuiltPhase
	// GetMoveCount 移動回数を取得する
	GetMoveCount() int
	// GetStockCount 山札の残り枚数を取得する
	GetStockCount() int
	// GetWaste 捨て札を取得する
	GetWaste() []*domain.Card
	// GetQuilt キルトを取得する（マス番号順、取り除いたマスは nil）
	GetQuilt() [domain.CrazyQuiltCells]*domain.Card
	// GetFoundation 基礎札を取得する
	GetFoundation() [domain.CrazyQuiltFoundationCnt][]*domain.Card
	// AllFaceUp 全カードが表向きかを返す
	AllFaceUp() bool
	// IsStalemate 手詰まり状態を取得する
	IsStalemate() bool
	// UndoToEscape 手詰まりから抜けるために必要なアンドゥ回数を取得する
	UndoToEscape() int
}
