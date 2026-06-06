//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EasthavenInteractorIF イーストヘイブンインタラクターインタフェース
type EasthavenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal ストックからタブローに配る
	Deal() string
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

// EasthavenInteractor イーストヘイブンインタラクタークラス
type EasthavenInteractor struct {
	GameBase[interfaces.EasthavenGame]
	ep presenter.EasthavenPresenter
	solitaireActions[interfaces.EasthavenGame]
}

// NewEasthavenInteractor コンストラクタ
func NewEasthavenInteractor(e interfaces.EasthavenGame, ep presenter.EasthavenPresenter) *EasthavenInteractor {
	mustNotNil("EasthavenInteractor", map[string]any{"e": e, "ep": ep})
	return &EasthavenInteractor{
		GameBase:         GameBase[interfaces.EasthavenGame]{Game: e},
		ep:               ep,
		solitaireActions: newSolitaireActions[interfaces.EasthavenGame](e, ep),
	}
}

// Reset ゲーム初期化
func (ei *EasthavenInteractor) Reset() string {
	return runAndPresent(ei.Game, ei.ep, ei.Game.Reset)
}

// Deal ストックからタブローに配る
func (ei *EasthavenInteractor) Deal() string {
	return execAndPresent(ei.Game, ei.ep, ei.Game.Deal)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ei *EasthavenInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ei.Game, ei.ep, func() error {
		return ei.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ei *EasthavenInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ei.Game, ei.ep, func() error { return ei.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (ei *EasthavenInteractor) Hint() string {
	return ei.ep.HintOutput(ei.Game)
}

// ActionLog 棋譜を出力する
func (ei *EasthavenInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// RestoreEasthavenInteractor deserialises JSON into an EasthavenInteractor.
func RestoreEasthavenInteractor(data []byte, ep presenter.EasthavenPresenter) (*EasthavenInteractor, error) {
	return restoreAndBuild[domain.Easthaven](data, func(g *domain.Easthaven) *EasthavenInteractor {
		return &EasthavenInteractor{
			GameBase:         GameBase[interfaces.EasthavenGame]{Game: g},
			ep:               ep,
			solitaireActions: newSolitaireActions[interfaces.EasthavenGame](g, ep),
		}
	})
}
