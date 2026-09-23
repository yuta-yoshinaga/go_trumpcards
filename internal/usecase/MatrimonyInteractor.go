//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MatrimonyInteractorIF マトリモニー インタラクターインタフェース
type MatrimonyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(pile int) string
	// MoveWasteToFoundation 捨て札から基礎札へ移動
	MoveWasteToFoundation() string
	// MoveWasteToTableau 捨て札からタブローへ移動
	MoveWasteToTableau(pile int) string
	// MoveStockToTableau 山札から空き山へ直接置く
	MoveStockToTableau(pile int) string
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

// MatrimonyInteractor マトリモニー インタラクタークラス
type MatrimonyInteractor struct {
	GameBase[interfaces.MatrimonyGame]
	cp presenter.MatrimonyPresenter
	solitaireActions[interfaces.MatrimonyGame]
}

// NewMatrimonyInteractor コンストラクタ
func NewMatrimonyInteractor(c interfaces.MatrimonyGame, cp presenter.MatrimonyPresenter) *MatrimonyInteractor {
	mustNotNil("MatrimonyInteractor", map[string]any{"c": c, "cp": cp})
	return &MatrimonyInteractor{
		GameBase:         GameBase[interfaces.MatrimonyGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.MatrimonyGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *MatrimonyInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *MatrimonyInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *MatrimonyInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *MatrimonyInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ci *MatrimonyInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(pile) })
}

// MoveStockToTableau 山札から空き山へ直接置く
func (ci *MatrimonyInteractor) MoveStockToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveStockToTableau(pile) })
}

// Hint ヒント取得
func (ci *MatrimonyInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *MatrimonyInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreMatrimonyInteractor deserialises JSON into a MatrimonyInteractor.
func RestoreMatrimonyInteractor(data []byte, cp presenter.MatrimonyPresenter) (*MatrimonyInteractor, error) {
	return restoreAndBuild[domain.Matrimony](data, func(g *domain.Matrimony) *MatrimonyInteractor {
		return &MatrimonyInteractor{
			GameBase:         GameBase[interfaces.MatrimonyGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.MatrimonyGame](g, cp),
		}
	})
}
