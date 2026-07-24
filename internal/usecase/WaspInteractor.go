//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WaspInteractorIF ワスプインタラクターインタフェース
type WaspInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal ストックからタブローに配る
	Deal() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// LegalMoves 指定列のトップカードの合法な移動先を出力する
	LegalMoves(col int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// WaspInteractor ワスプインタラクタークラス
type WaspInteractor struct {
	GameBase[interfaces.WaspGame]
	sp presenter.WaspPresenter
	solitaireActions[interfaces.WaspGame]
}

// NewWaspInteractor コンストラクタ
func NewWaspInteractor(s interfaces.WaspGame, sp presenter.WaspPresenter) *WaspInteractor {
	mustNotNil("WaspInteractor", map[string]any{"s": s, "sp": sp})
	return &WaspInteractor{
		GameBase:         GameBase[interfaces.WaspGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.WaspGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *WaspInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Deal ストックからタブローに配る
func (si *WaspInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.Deal)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (si *WaspInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error {
		return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// Hint ヒント取得
func (si *WaspInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// LegalMoves 指定列のトップカードの合法な移動先を出力する
func (si *WaspInteractor) LegalMoves(col int) string {
	return si.sp.LegalMovesOutput(si.Game, col)
}

// ActionLog 棋譜を出力する
func (si *WaspInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreWaspInteractor deserialises JSON into a WaspInteractor.
func RestoreWaspInteractor(data []byte, sp presenter.WaspPresenter) (*WaspInteractor, error) {
	return restoreAndBuild[domain.Wasp](data, func(g *domain.Wasp) *WaspInteractor {
		return &WaspInteractor{
			GameBase:         GameBase[interfaces.WaspGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.WaspGame](g, sp),
		}
	})
}
