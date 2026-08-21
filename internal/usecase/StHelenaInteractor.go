//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// StHelenaInteractorIF セント・ヘレナ・ソリティアのインタラクタインタフェース。
type StHelenaInteractorIF interface {
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

// StHelenaInteractor セント・ヘレナ・ソリティアのインタラクタ。
type StHelenaInteractor struct {
	GameBase[interfaces.StHelenaGame]
	cp presenter.StHelenaPresenter
	solitaireActions[interfaces.StHelenaGame]
}

// NewStHelenaInteractor コンストラクタ。
func NewStHelenaInteractor(cr interfaces.StHelenaGame, cp presenter.StHelenaPresenter) *StHelenaInteractor {
	mustNotNil("StHelenaInteractor", map[string]any{"cr": cr, "cp": cp})
	return &StHelenaInteractor{
		GameBase:         GameBase[interfaces.StHelenaGame]{Game: cr},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.StHelenaGame](cr, cp),
	}
}

// Reset ゲーム初期化。
func (ci *StHelenaInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// MoveTableauToTableau タブロー間でカードを移動。
func (ci *StHelenaInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromCol, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションへ移動。
func (ci *StHelenaInteractor) MoveTableauToFoundation(fromCol, foundationIdx int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(fromCol, foundationIdx) })
}

// Redeal 再配りを実行。
func (ci *StHelenaInteractor) Redeal() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Redeal)
}

// Hint ヒント取得。
func (ci *StHelenaInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜出力。
func (ci *StHelenaInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreStHelenaInteractor JSON から StHelenaInteractor を復元する。
func RestoreStHelenaInteractor(data []byte, cp presenter.StHelenaPresenter) (*StHelenaInteractor, error) {
	return restoreAndBuild[domain.StHelena](data, func(g *domain.StHelena) *StHelenaInteractor {
		return &StHelenaInteractor{
			GameBase:         GameBase[interfaces.StHelenaGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.StHelenaGame](g, cp),
		}
	})
}
