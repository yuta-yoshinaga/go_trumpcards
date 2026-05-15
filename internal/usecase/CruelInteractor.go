package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CruelInteractorIF クルーエルインタラクターインタフェース
type CruelInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, toCol int) string
	// MoveTableauToFoundation タブローからファウンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// Shift タブローを再構築
	Shift() string
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

// CruelInteractor クルーエルインタラクタークラス
type CruelInteractor struct {
	GameBase[interfaces.CruelGame]
	cp presenter.CruelPresenter
	solitaireActions[interfaces.CruelGame]
}

// NewCruelInteractor コンストラクタ
func NewCruelInteractor(c interfaces.CruelGame, cp presenter.CruelPresenter) *CruelInteractor {
	mustNotNil("CruelInteractor", map[string]any{"c": c, "cp": cp})
	return &CruelInteractor{
		GameBase:         GameBase[interfaces.CruelGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.CruelGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *CruelInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ci *CruelInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(ci.Game, ci.cp, func() error {
		return ci.Game.MoveTableauToTableau(fromCol, toCol)
	})
}

// MoveTableauToFoundation タブローからファウンデーションにカードを移動
func (ci *CruelInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(col) })
}

// Shift タブローを左から12列×4枚に配り直す
func (ci *CruelInteractor) Shift() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Shift)
}

// Hint ヒント取得
func (ci *CruelInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CruelInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCruelInteractor deserialises JSON into a CruelInteractor.
func RestoreCruelInteractor(data []byte, cp presenter.CruelPresenter) (*CruelInteractor, error) {
	return restoreAndBuild[domain.Cruel](data, func(g *domain.Cruel) *CruelInteractor {
		return &CruelInteractor{
			GameBase:         GameBase[interfaces.CruelGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.CruelGame](g, cp),
		}
	})
}
