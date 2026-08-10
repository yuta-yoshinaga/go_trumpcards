//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ColoradoInteractorIF コロラド インタラクターインタフェース
type ColoradoInteractorIF interface {
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

// ColoradoInteractor コロラド インタラクタークラス
type ColoradoInteractor struct {
	GameBase[interfaces.ColoradoGame]
	cp presenter.ColoradoPresenter
	solitaireActions[interfaces.ColoradoGame]
}

// NewColoradoInteractor コンストラクタ
func NewColoradoInteractor(c interfaces.ColoradoGame, cp presenter.ColoradoPresenter) *ColoradoInteractor {
	mustNotNil("ColoradoInteractor", map[string]any{"c": c, "cp": cp})
	return &ColoradoInteractor{
		GameBase:         GameBase[interfaces.ColoradoGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.ColoradoGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *ColoradoInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *ColoradoInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *ColoradoInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *ColoradoInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ci *ColoradoInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(pile) })
}

// MoveStockToTableau 山札から空き山へ直接置く
func (ci *ColoradoInteractor) MoveStockToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveStockToTableau(pile) })
}

// Hint ヒント取得
func (ci *ColoradoInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *ColoradoInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreColoradoInteractor deserialises JSON into a ColoradoInteractor.
func RestoreColoradoInteractor(data []byte, cp presenter.ColoradoPresenter) (*ColoradoInteractor, error) {
	return restoreAndBuild[domain.Colorado](data, func(g *domain.Colorado) *ColoradoInteractor {
		return &ColoradoInteractor{
			GameBase:         GameBase[interfaces.ColoradoGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.ColoradoGame](g, cp),
		}
	})
}
