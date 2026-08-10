//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DiplomatInteractorIF ディプロマット インタラクターインタフェース
type DiplomatInteractorIF interface {
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

// DiplomatInteractor ディプロマット インタラクタークラス
type DiplomatInteractor struct {
	GameBase[interfaces.DiplomatGame]
	cp presenter.DiplomatPresenter
	solitaireActions[interfaces.DiplomatGame]
}

// NewDiplomatInteractor コンストラクタ
func NewDiplomatInteractor(c interfaces.DiplomatGame, cp presenter.DiplomatPresenter) *DiplomatInteractor {
	mustNotNil("DiplomatInteractor", map[string]any{"c": c, "cp": cp})
	return &DiplomatInteractor{
		GameBase:         GameBase[interfaces.DiplomatGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.DiplomatGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *DiplomatInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *DiplomatInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *DiplomatInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveTableauToTableau タブロー間で移動
func (ci *DiplomatInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromPile, toPile) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *DiplomatInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ci *DiplomatInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveWasteToTableau(pile) })
}

// Hint ヒント取得
func (ci *DiplomatInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *DiplomatInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreDiplomatInteractor deserialises JSON into a DiplomatInteractor.
func RestoreDiplomatInteractor(data []byte, cp presenter.DiplomatPresenter) (*DiplomatInteractor, error) {
	return restoreAndBuild[domain.Diplomat](data, func(g *domain.Diplomat) *DiplomatInteractor {
		return &DiplomatInteractor{
			GameBase:         GameBase[interfaces.DiplomatGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.DiplomatGame](g, cp),
		}
	})
}
