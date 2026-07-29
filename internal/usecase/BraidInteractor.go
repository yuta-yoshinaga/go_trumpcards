//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BraidInteractorIF ブレイド インタラクターインタフェース
type BraidInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// ChooseDirection 基礎札を積む向きを決める
	ChooseDirection(ascending bool) string
	// MoveBraidToFoundation ブレイドから基礎札へ移動
	MoveBraidToFoundation() string
	// MoveFieldToFoundation ブレイド札から基礎札へ移動
	MoveFieldToFoundation(idx int) string
	// MoveHelperToFoundation ヘルパーから基礎札へ移動
	MoveHelperToFoundation(idx int) string
	// MoveWasteToFoundation 捨て札から基礎札へ移動
	MoveWasteToFoundation() string
	// MoveWasteToHelper 捨て札からヘルパーへ移動
	MoveWasteToHelper(idx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// BraidInteractor ブレイド インタラクタークラス
type BraidInteractor struct {
	GameBase[interfaces.BraidGame]
	bp presenter.BraidPresenter
	solitaireActions[interfaces.BraidGame]
}

// NewBraidInteractor コンストラクタ
func NewBraidInteractor(b interfaces.BraidGame, bp presenter.BraidPresenter) *BraidInteractor {
	mustNotNil("BraidInteractor", map[string]any{"b": b, "bp": bp})
	return &BraidInteractor{
		GameBase:         GameBase[interfaces.BraidGame]{Game: b},
		bp:               bp,
		solitaireActions: newSolitaireActions[interfaces.BraidGame](b, bp),
	}
}

// Reset ゲーム初期化
func (bi *BraidInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bp, bi.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (bi *BraidInteractor) Draw() string {
	return execAndPresent(bi.Game, bi.bp, bi.Game.Draw)
}

// ChooseDirection 基礎札を積む向きを決める
func (bi *BraidInteractor) ChooseDirection(ascending bool) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.ChooseDirection(ascending) })
}

// MoveBraidToFoundation ブレイドから基礎札へ移動
func (bi *BraidInteractor) MoveBraidToFoundation() string {
	return execAndPresent(bi.Game, bi.bp, bi.Game.MoveBraidToFoundation)
}

// MoveFieldToFoundation ブレイド札から基礎札へ移動
func (bi *BraidInteractor) MoveFieldToFoundation(idx int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveFieldToFoundation(idx) })
}

// MoveHelperToFoundation ヘルパーから基礎札へ移動
func (bi *BraidInteractor) MoveHelperToFoundation(idx int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveHelperToFoundation(idx) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (bi *BraidInteractor) MoveWasteToFoundation() string {
	return execAndPresent(bi.Game, bi.bp, bi.Game.MoveWasteToFoundation)
}

// MoveWasteToHelper 捨て札からヘルパーへ移動
func (bi *BraidInteractor) MoveWasteToHelper(idx int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveWasteToHelper(idx) })
}

// Hint ヒント取得
func (bi *BraidInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BraidInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// RestoreBraidInteractor deserialises JSON into a BraidInteractor.
func RestoreBraidInteractor(data []byte, bp presenter.BraidPresenter) (*BraidInteractor, error) {
	return restoreAndBuild[domain.Braid](data, func(g *domain.Braid) *BraidInteractor {
		return &BraidInteractor{
			GameBase:         GameBase[interfaces.BraidGame]{Game: g},
			bp:               bp,
			solitaireActions: newSolitaireActions[interfaces.BraidGame](g, bp),
		}
	})
}
