//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WillOTheWispInteractorIF ウィル・オ・ザ・ウィスプインタラクターインタフェース
type WillOTheWispInteractorIF interface {
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

// WillOTheWispInteractor ウィル・オ・ザ・ウィスプインタラクタークラス
type WillOTheWispInteractor struct {
	GameBase[interfaces.WillOTheWispGame]
	sp presenter.WillOTheWispPresenter
	solitaireActions[interfaces.WillOTheWispGame]
}

// NewWillOTheWispInteractor コンストラクタ
func NewWillOTheWispInteractor(s interfaces.WillOTheWispGame, sp presenter.WillOTheWispPresenter) *WillOTheWispInteractor {
	mustNotNil("WillOTheWispInteractor", map[string]any{"s": s, "sp": sp})
	return &WillOTheWispInteractor{
		GameBase:         GameBase[interfaces.WillOTheWispGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.WillOTheWispGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *WillOTheWispInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Deal ストックからタブローに配る
func (si *WillOTheWispInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.Deal)
}

// MoveTableauToTableau タブロー間でカードを移動
func (si *WillOTheWispInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// Hint ヒント取得
func (si *WillOTheWispInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *WillOTheWispInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreWillOTheWispInteractor deserialises JSON into a WillOTheWispInteractor.
func RestoreWillOTheWispInteractor(data []byte, sp presenter.WillOTheWispPresenter) (*WillOTheWispInteractor, error) {
	return restoreAndBuild[domain.WillOTheWisp](data, func(g *domain.WillOTheWisp) *WillOTheWispInteractor {
		return &WillOTheWispInteractor{
			GameBase:         GameBase[interfaces.WillOTheWispGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.WillOTheWispGame](g, sp),
		}
	})
}
