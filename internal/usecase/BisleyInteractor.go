//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BisleyInteractorIF ビズリー インタラクターインタフェース
type BisleyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, toCol int) string
	// MoveTableauToAceFoundation タブローから昇順基礎札にカードを移動
	MoveTableauToAceFoundation(col int) string
	// MoveTableauToKingFoundation タブローから降順基礎札にカードを移動
	MoveTableauToKingFoundation(col int) string
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

// BisleyInteractor ビズリー インタラクタークラス
type BisleyInteractor struct {
	GameBase[interfaces.BisleyGame]
	bp presenter.BisleyPresenter
	solitaireActions[interfaces.BisleyGame]
}

// NewBisleyInteractor コンストラクタ
func NewBisleyInteractor(b interfaces.BisleyGame, bp presenter.BisleyPresenter) *BisleyInteractor {
	mustNotNil("BisleyInteractor", map[string]any{"b": b, "bp": bp})
	return &BisleyInteractor{
		GameBase:         GameBase[interfaces.BisleyGame]{Game: b},
		bp:               bp,
		solitaireActions: newSolitaireActions[interfaces.BisleyGame](b, bp),
	}
}

// Reset ゲーム初期化
func (bi *BisleyInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *BisleyInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveTableauToTableau(fromCol, toCol) })
}

// MoveTableauToAceFoundation タブローから昇順基礎札にカードを移動
func (bi *BisleyInteractor) MoveTableauToAceFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveTableauToAceFoundation(col) })
}

// MoveTableauToKingFoundation タブローから降順基礎札にカードを移動
func (bi *BisleyInteractor) MoveTableauToKingFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.MoveTableauToKingFoundation(col) })
}

// Hint ヒント取得
func (bi *BisleyInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BisleyInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// RestoreBisleyInteractor deserialises JSON into a BisleyInteractor.
func RestoreBisleyInteractor(data []byte, bp presenter.BisleyPresenter) (*BisleyInteractor, error) {
	return restoreAndBuild[domain.Bisley](data, func(g *domain.Bisley) *BisleyInteractor {
		return &BisleyInteractor{
			GameBase:         GameBase[interfaces.BisleyGame]{Game: g},
			bp:               bp,
			solitaireActions: newSolitaireActions[interfaces.BisleyGame](g, bp),
		}
	})
}
