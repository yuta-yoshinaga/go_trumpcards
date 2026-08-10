//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FourSeasonsGame フォーシーズンズゲームインタフェース
type FourSeasonsGame interface {
	BaseGame
	// GetGameEndFlag reports whether the game has left the playing phase.
	GetGameEndFlag() bool
	// Reset ゲームを初期化する
	Reset()
	// Draw 山札からカードをめくる
	Draw() error
	// MoveWasteToTableau ウェイストからタブローに移動
	MoveWasteToTableau(col int) error
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation(fIdx int) error
	// MoveTableauToTableau タブロー間で移動（最上段の1枚のみ）
	MoveTableauToTableau(fromCol, toCol int) error
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col, fIdx int) error
	// GiveUp ギブアップ
	GiveUp()
	// GetHint ヒント取得
	GetHint() *domain.FourSeasonsHint
	// AutoComplete 自動完了
	AutoComplete() error
	// Undo アンドゥ
	Undo() error
	// CanUndo アンドゥ可能か
	CanUndo() bool
	// UndoN n回アンドゥ
	UndoN(n int) error

	// GetPhase フェーズ取得
	GetPhase() domain.FourSeasonsPhase
	// GetMoveCount 移動回数
	GetMoveCount() int
	// GetStockCount ストック残枚数
	GetStockCount() int
	// GetWaste ウェイスト取得
	GetWaste() []*domain.Card
	// GetTableau タブロー（十字）取得
	GetTableau() [domain.FourSeasonsTableauCnt][]*domain.Card
	// GetFoundations ファンデーション（四隅）取得
	GetFoundations() [domain.FourSeasonsFoundationCnt][]*domain.Card
	// GetBaseRank ベースランク取得。配りごとに変わり、すべての配置判定がこれに乗る。
	GetBaseRank() int
}
