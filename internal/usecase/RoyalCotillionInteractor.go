//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RoyalCotillionInteractorIF ロイヤルコティヨン インタラクターインタフェース
type RoyalCotillionInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(pile int) string
	// MoveReserveToFoundation リザーブから基礎札へ移動
	MoveReserveToFoundation(pile int) string
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

// RoyalCotillionInteractor ロイヤルコティヨン インタラクタークラス
type RoyalCotillionInteractor struct {
	GameBase[interfaces.RoyalCotillionGame]
	cp presenter.RoyalCotillionPresenter
	solitaireActions[interfaces.RoyalCotillionGame]
}

// NewRoyalCotillionInteractor コンストラクタ
func NewRoyalCotillionInteractor(c interfaces.RoyalCotillionGame, cp presenter.RoyalCotillionPresenter) *RoyalCotillionInteractor {
	mustNotNil("RoyalCotillionInteractor", map[string]any{"c": c, "cp": cp})
	return &RoyalCotillionInteractor{
		GameBase:         GameBase[interfaces.RoyalCotillionGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.RoyalCotillionGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *RoyalCotillionInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *RoyalCotillionInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *RoyalCotillionInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveReserveToFoundation リザーブから基礎札へ移動
func (ci *RoyalCotillionInteractor) MoveReserveToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveReserveToFoundation(pile) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *RoyalCotillionInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ci *RoyalCotillionInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(pile) })
}

// MoveStockToTableau 山札から空き山へ直接置く
func (ci *RoyalCotillionInteractor) MoveStockToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveStockToTableau(pile) })
}

// Hint ヒント取得
func (ci *RoyalCotillionInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *RoyalCotillionInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreRoyalCotillionInteractor deserialises JSON into a RoyalCotillionInteractor.
func RestoreRoyalCotillionInteractor(data []byte, cp presenter.RoyalCotillionPresenter) (*RoyalCotillionInteractor, error) {
	return restoreAndBuild[domain.RoyalCotillion](data, func(g *domain.RoyalCotillion) *RoyalCotillionInteractor {
		return &RoyalCotillionInteractor{
			GameBase:         GameBase[interfaces.RoyalCotillionGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.RoyalCotillionGame](g, cp),
		}
	})
}
