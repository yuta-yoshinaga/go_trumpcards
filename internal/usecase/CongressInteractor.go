//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CongressInteractorIF コングレス インタラクターインタフェース
type CongressInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(pile int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromPile, toPile int) string
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

// CongressInteractor コングレス インタラクタークラス
type CongressInteractor struct {
	GameBase[interfaces.CongressGame]
	cp presenter.CongressPresenter
	solitaireActions[interfaces.CongressGame]
}

// NewCongressInteractor コンストラクタ
func NewCongressInteractor(c interfaces.CongressGame, cp presenter.CongressPresenter) *CongressInteractor {
	mustNotNil("CongressInteractor", map[string]any{"c": c, "cp": cp})
	return &CongressInteractor{
		GameBase:         GameBase[interfaces.CongressGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.CongressGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *CongressInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *CongressInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *CongressInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveTableauToTableau タブロー間で移動
func (ci *CongressInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromPile, toPile) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *CongressInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ci *CongressInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(pile) })
}

// MoveStockToTableau 山札から空き山へ直接置く
func (ci *CongressInteractor) MoveStockToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveStockToTableau(pile) })
}

// Hint ヒント取得
func (ci *CongressInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CongressInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCongressInteractor deserialises JSON into a CongressInteractor.
func RestoreCongressInteractor(data []byte, cp presenter.CongressPresenter) (*CongressInteractor, error) {
	return restoreAndBuild[domain.Congress](data, func(g *domain.Congress) *CongressInteractor {
		return &CongressInteractor{
			GameBase:         GameBase[interfaces.CongressGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.CongressGame](g, cp),
		}
	})
}
