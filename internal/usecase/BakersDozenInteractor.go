//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BakersDozenInteractorIF ベーカーズダズンインタラクターインタフェース
type BakersDozenInteractorIF interface {
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
	// Targets 列 col の一番下の札を置ける先を一覧する
	Targets(col int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// BakersDozenInteractor ベーカーズダズンインタラクタークラス
type BakersDozenInteractor struct {
	GameBase[interfaces.BakersDozenGame]
	bdp presenter.BakersDozenPresenter
	solitaireActions[interfaces.BakersDozenGame]
}

// NewBakersDozenInteractor コンストラクタ
func NewBakersDozenInteractor(bd interfaces.BakersDozenGame, bdp presenter.BakersDozenPresenter) *BakersDozenInteractor {
	mustNotNil("BakersDozenInteractor", map[string]any{"bd": bd, "bdp": bdp})
	return &BakersDozenInteractor{
		GameBase:         GameBase[interfaces.BakersDozenGame]{Game: bd},
		bdp:              bdp,
		solitaireActions: newSolitaireActions[interfaces.BakersDozenGame](bd, bdp),
	}
}

// Reset ゲーム初期化
func (bi *BakersDozenInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bdp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *BakersDozenInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bdp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *BakersDozenInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bdp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *BakersDozenInteractor) Hint() string {
	return bi.bdp.HintOutput(bi.Game)
}

// Targets 列 col の一番下の札を置ける先を一覧する
func (bi *BakersDozenInteractor) Targets(col int) string {
	return bi.bdp.TargetsOutput(bi.Game, col)
}

// ActionLog 棋譜を出力する
func (bi *BakersDozenInteractor) ActionLog() string {
	return bi.bdp.ActionLogOutput(bi.Game)
}

// RestoreBakersDozenInteractor deserialises JSON into a BakersDozenInteractor.
func RestoreBakersDozenInteractor(data []byte, bdp presenter.BakersDozenPresenter) (*BakersDozenInteractor, error) {
	return restoreAndBuild[domain.BakersDozen](data, func(g *domain.BakersDozen) *BakersDozenInteractor {
		return &BakersDozenInteractor{
			GameBase:         GameBase[interfaces.BakersDozenGame]{Game: g},
			bdp:              bdp,
			solitaireActions: newSolitaireActions[interfaces.BakersDozenGame](g, bdp),
		}
	})
}
