//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TerraceInteractorIF テラス インタラクターインタフェース
type TerraceInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveReserveToFoundation テラスから基礎札へ移動
	MoveReserveToFoundation() string
	// MoveWasteToFoundation 捨て札から基礎札へ移動
	MoveWasteToFoundation() string
	// MoveWasteToTableau 捨て札からタブローへ移動
	MoveWasteToTableau(pile int) string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(pile int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromPile, toPile int) string
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

// TerraceInteractor テラス インタラクタークラス
type TerraceInteractor struct {
	GameBase[interfaces.TerraceGame]
	tp presenter.TerracePresenter
	solitaireActions[interfaces.TerraceGame]
}

// NewTerraceInteractor コンストラクタ
func NewTerraceInteractor(t interfaces.TerraceGame, tp presenter.TerracePresenter) *TerraceInteractor {
	mustNotNil("TerraceInteractor", map[string]any{"t": t, "tp": tp})
	return &TerraceInteractor{
		GameBase:         GameBase[interfaces.TerraceGame]{Game: t},
		tp:               tp,
		solitaireActions: newSolitaireActions[interfaces.TerraceGame](t, tp),
	}
}

// Reset ゲーム初期化
func (ti *TerraceInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ti *TerraceInteractor) Draw() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Draw)
}

// MoveReserveToFoundation テラスから基礎札へ移動
func (ti *TerraceInteractor) MoveReserveToFoundation() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.MoveReserveToFoundation)
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ti *TerraceInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ti *TerraceInteractor) MoveWasteToTableau(pile int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.MoveWasteToTableau(pile) })
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ti *TerraceInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.MoveTableauToFoundation(pile) })
}

// MoveTableauToTableau タブロー間で移動
func (ti *TerraceInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.MoveTableauToTableau(fromPile, toPile) })
}

// Hint ヒント取得
func (ti *TerraceInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TerraceInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// RestoreTerraceInteractor deserialises JSON into a TerraceInteractor.
func RestoreTerraceInteractor(data []byte, tp presenter.TerracePresenter) (*TerraceInteractor, error) {
	return restoreAndBuild[domain.Terrace](data, func(g *domain.Terrace) *TerraceInteractor {
		return &TerraceInteractor{
			GameBase:         GameBase[interfaces.TerraceGame]{Game: g},
			tp:               tp,
			solitaireActions: newSolitaireActions[interfaces.TerraceGame](g, tp),
		}
	})
}
