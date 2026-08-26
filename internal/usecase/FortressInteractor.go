//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FortressInteractorIF フォートレスインタラクターインタフェース
type FortressInteractorIF interface {
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

// FortressInteractor フォートレスインタラクタークラス
type FortressInteractor struct {
	GameBase[interfaces.FortressGame]
	bcp presenter.FortressPresenter
	solitaireActions[interfaces.FortressGame]
}

// NewFortressInteractor コンストラクタ
func NewFortressInteractor(bc interfaces.FortressGame, bcp presenter.FortressPresenter) *FortressInteractor {
	mustNotNil("FortressInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &FortressInteractor{
		GameBase:         GameBase[interfaces.FortressGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.FortressGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *FortressInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *FortressInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *FortressInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *FortressInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *FortressInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreFortressInteractor deserialises JSON into a FortressInteractor.
func RestoreFortressInteractor(data []byte, bcp presenter.FortressPresenter) (*FortressInteractor, error) {
	return restoreAndBuild[domain.Fortress](data, func(g *domain.Fortress) *FortressInteractor {
		return &FortressInteractor{
			GameBase:         GameBase[interfaces.FortressGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.FortressGame](g, bcp),
		}
	})
}
