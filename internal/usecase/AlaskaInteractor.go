//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AlaskaInteractorIF アラスカインタラクターインタフェース
type AlaskaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
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

// AlaskaInteractor アラスカインタラクタークラス
type AlaskaInteractor struct {
	GameBase[interfaces.AlaskaGame]
	rp presenter.AlaskaPresenter
	solitaireActions[interfaces.AlaskaGame]
}

// NewAlaskaInteractor コンストラクタ
func NewAlaskaInteractor(r interfaces.AlaskaGame, rp presenter.AlaskaPresenter) *AlaskaInteractor {
	mustNotNil("AlaskaInteractor", map[string]any{"r": r, "rp": rp})
	return &AlaskaInteractor{
		GameBase:         GameBase[interfaces.AlaskaGame]{Game: r},
		rp:               rp,
		solitaireActions: newSolitaireActions[interfaces.AlaskaGame](r, rp),
	}
}

// Reset ゲーム初期化
func (ri *AlaskaInteractor) Reset() string {
	return runAndPresent(ri.Game, ri.rp, ri.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ri *AlaskaInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ri.Game, ri.rp, func() error {
		return ri.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ri *AlaskaInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ri.Game, ri.rp, func() error { return ri.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (ri *AlaskaInteractor) Hint() string {
	return ri.rp.HintOutput(ri.Game)
}

// ActionLog 棋譜を出力する
func (ri *AlaskaInteractor) ActionLog() string {
	return ri.rp.ActionLogOutput(ri.Game)
}

// RestoreAlaskaInteractor deserialises JSON into a AlaskaInteractor.
func RestoreAlaskaInteractor(data []byte, rp presenter.AlaskaPresenter) (*AlaskaInteractor, error) {
	return restoreAndBuild[domain.Alaska](data, func(g *domain.Alaska) *AlaskaInteractor {
		return &AlaskaInteractor{
			GameBase:         GameBase[interfaces.AlaskaGame]{Game: g},
			rp:               rp,
			solitaireActions: newSolitaireActions[interfaces.AlaskaGame](g, rp),
		}
	})
}
