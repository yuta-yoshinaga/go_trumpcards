package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CrescentInteractorIF クレセント・ソリティアのインタラクタインタフェース。
type CrescentInteractorIF interface {
	// Snapshot KV 永続化用のシリアライズ。
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// MoveTableauToTableau タブロー間でカードを移動。
	MoveTableauToTableau(fromCol, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションへ移動。
	MoveTableauToFoundation(fromCol, foundationIdx int) string
	// Redeal 再配り (シャッフル) を実行。
	Redeal() string
	// GiveUp ギブアップ。
	GiveUp() string
	// Hint ヒント取得。
	Hint() string
	// AutoComplete オートコンプリート。
	AutoComplete() string
	// ActionLog 棋譜出力。
	ActionLog() string
	// Undo アンドゥ。
	Undo() string
	// UndoN n 回連続アンドゥ。
	UndoN(n int) string
}

// CrescentInteractor クレセント・ソリティアのインタラクタ。
type CrescentInteractor struct {
	GameBase[interfaces.CrescentGame]
	cp presenter.CrescentPresenter
	solitaireActions[interfaces.CrescentGame]
}

// NewCrescentInteractor コンストラクタ。
func NewCrescentInteractor(cr interfaces.CrescentGame, cp presenter.CrescentPresenter) *CrescentInteractor {
	mustNotNil("CrescentInteractor", map[string]any{"cr": cr, "cp": cp})
	return &CrescentInteractor{
		GameBase:         GameBase[interfaces.CrescentGame]{Game: cr},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.CrescentGame](cr, cp),
	}
}

// Reset ゲーム初期化。
func (ci *CrescentInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// MoveTableauToTableau タブロー間でカードを移動。
func (ci *CrescentInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromCol, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションへ移動。
func (ci *CrescentInteractor) MoveTableauToFoundation(fromCol, foundationIdx int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(fromCol, foundationIdx) })
}

// Redeal 再配りを実行。
func (ci *CrescentInteractor) Redeal() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Redeal)
}

// Hint ヒント取得。
func (ci *CrescentInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜出力。
func (ci *CrescentInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCrescentInteractor JSON から CrescentInteractor を復元する。
func RestoreCrescentInteractor(data []byte, cp presenter.CrescentPresenter) (*CrescentInteractor, error) {
	return restoreAndBuild[domain.Crescent](data, func(g *domain.Crescent) *CrescentInteractor {
		return &CrescentInteractor{
			GameBase:         GameBase[interfaces.CrescentGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.CrescentGame](g, cp),
		}
	})
}
