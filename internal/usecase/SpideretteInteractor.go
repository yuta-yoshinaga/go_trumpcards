//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpideretteInteractorIF スパイダレットインタラクターインタフェース
type SpideretteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal ストックからタブローに配る
	Deal() string
	// MoveTableauToTableau タブロー間でカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
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

// SpideretteInteractor スパイダレットインタラクタークラス
type SpideretteInteractor struct {
	GameBase[interfaces.SpideretteGame]
	sp presenter.SpiderettePresenter
	solitaireActions[interfaces.SpideretteGame]
}

// NewSpideretteInteractor コンストラクタ
func NewSpideretteInteractor(s interfaces.SpideretteGame, sp presenter.SpiderettePresenter) *SpideretteInteractor {
	mustNotNil("SpideretteInteractor", map[string]any{"s": s, "sp": sp})
	return &SpideretteInteractor{
		GameBase:         GameBase[interfaces.SpideretteGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.SpideretteGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *SpideretteInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Deal ストックからタブローに配る
func (si *SpideretteInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.Deal)
}

// MoveTableauToTableau タブロー間でカードを移動
func (si *SpideretteInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// Hint ヒント取得
func (si *SpideretteInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SpideretteInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSpideretteInteractor deserialises JSON into a SpideretteInteractor.
func RestoreSpideretteInteractor(data []byte, sp presenter.SpiderettePresenter) (*SpideretteInteractor, error) {
	return restoreAndBuild[domain.Spiderette](data, func(g *domain.Spiderette) *SpideretteInteractor {
		return &SpideretteInteractor{
			GameBase:         GameBase[interfaces.SpideretteGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.SpideretteGame](g, sp),
		}
	})
}
