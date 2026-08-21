//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ShamrocksGame シャムロックス (ファン・ソリティア) のゲームインタフェース。
type ShamrocksGame interface {
	BaseGame
	// GetGameEndFlag プレイ中でなければ true。
	GetGameEndFlag() bool
	// Reset ゲームを初期化する。
	Reset()
	// MoveFanToFan 扇から扇へカードを移す。
	MoveFanToFan(from, to int) error
	// MoveFanToFoundation 扇からファウンデーションへ移す。
	MoveFanToFoundation(from int) error
	// GiveUp 投了する。
	GiveUp()
	// AutoComplete 出せるファウンデーション手を自動で出し切る。
	AutoComplete() error
	// Undo 直近の手を取り消す。
	Undo() error
	// CanUndo Undo 可能か。
	CanUndo() bool
	// UndoN n 回 Undo する。
	UndoN(n int) error
	// GetHint ヒントを取得する。
	GetHint() *domain.ShamrocksHint

	// GetPhase 現在のフェーズを取得する。
	GetPhase() domain.ShamrocksPhase
	// GetMoveCount 累計手数を取得する。
	GetMoveCount() int
	// HasAnyLegalMove 合法手が存在するかを返す (なければリディールが必要)。
	HasAnyLegalMove() bool
	// GetFans 扇の一覧を取得する。
	GetFans() [][]*domain.Card
	// GetFoundation ファウンデーションを取得する。
	GetFoundation() [domain.ShamrocksFoundationCnt][]*domain.Card
}
