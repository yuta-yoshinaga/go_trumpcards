//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SalicLawInteractorIF サリカ法典 インタラクターインタフェース
type SalicLawInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から 1 枚めくって今の列に置く（K なら次の列を開く）
	Draw() string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(pile int) string
	// MoveTableauToTableau 「K だけの列」へ移動
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

// SalicLawInteractor サリカ法典 インタラクタークラス
type SalicLawInteractor struct {
	GameBase[interfaces.SalicLawGame]
	cp presenter.SalicLawPresenter
	solitaireActions[interfaces.SalicLawGame]
}

// NewSalicLawInteractor コンストラクタ
func NewSalicLawInteractor(c interfaces.SalicLawGame, cp presenter.SalicLawPresenter) *SalicLawInteractor {
	mustNotNil("SalicLawInteractor", map[string]any{"c": c, "cp": cp})
	return &SalicLawInteractor{
		GameBase:         GameBase[interfaces.SalicLawGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.SalicLawGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *SalicLawInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から 1 枚めくって今の列に置く（K なら次の列を開く）
func (ci *SalicLawInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ci *SalicLawInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// MoveTableauToTableau 「K だけの列」へ移動
func (ci *SalicLawInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToTableau(fromPile, toPile) })
}

// Hint ヒント取得
func (ci *SalicLawInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *SalicLawInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreSalicLawInteractor deserialises JSON into a SalicLawInteractor.
func RestoreSalicLawInteractor(data []byte, cp presenter.SalicLawPresenter) (*SalicLawInteractor, error) {
	return restoreAndBuild[domain.SalicLaw](data, func(g *domain.SalicLaw) *SalicLawInteractor {
		return &SalicLawInteractor{
			GameBase:         GameBase[interfaces.SalicLawGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.SalicLawGame](g, cp),
		}
	})
}
