//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RussianSolitaireInteractorIF ロシアンソリティアインタラクターインタフェース
type RussianSolitaireInteractorIF interface {
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

// RussianSolitaireInteractor ロシアンソリティアインタラクタークラス
type RussianSolitaireInteractor struct {
	GameBase[interfaces.RussianSolitaireGame]
	rp presenter.RussianSolitairePresenter
	solitaireActions[interfaces.RussianSolitaireGame]
}

// NewRussianSolitaireInteractor コンストラクタ
func NewRussianSolitaireInteractor(r interfaces.RussianSolitaireGame, rp presenter.RussianSolitairePresenter) *RussianSolitaireInteractor {
	mustNotNil("RussianSolitaireInteractor", map[string]any{"r": r, "rp": rp})
	return &RussianSolitaireInteractor{
		GameBase:         GameBase[interfaces.RussianSolitaireGame]{Game: r},
		rp:               rp,
		solitaireActions: newSolitaireActions[interfaces.RussianSolitaireGame](r, rp),
	}
}

// Reset ゲーム初期化
func (ri *RussianSolitaireInteractor) Reset() string {
	return runAndPresent(ri.Game, ri.rp, ri.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ri *RussianSolitaireInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ri.Game, ri.rp, func() error {
		return ri.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ri *RussianSolitaireInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ri.Game, ri.rp, func() error { return ri.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (ri *RussianSolitaireInteractor) Hint() string {
	return ri.rp.HintOutput(ri.Game)
}

// ActionLog 棋譜を出力する
func (ri *RussianSolitaireInteractor) ActionLog() string {
	return ri.rp.ActionLogOutput(ri.Game)
}

// RestoreRussianSolitaireInteractor deserialises JSON into a RussianSolitaireInteractor.
func RestoreRussianSolitaireInteractor(data []byte, rp presenter.RussianSolitairePresenter) (*RussianSolitaireInteractor, error) {
	return restoreAndBuild[domain.RussianSolitaire](data, func(g *domain.RussianSolitaire) *RussianSolitaireInteractor {
		return &RussianSolitaireInteractor{
			GameBase:         GameBase[interfaces.RussianSolitaireGame]{Game: g},
			rp:               rp,
			solitaireActions: newSolitaireActions[interfaces.RussianSolitaireGame](g, rp),
		}
	})
}
