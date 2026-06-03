//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SeahavenTowersInteractorIF シーヘイブンタワーズインタラクターインタフェース
type SeahavenTowersInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// MoveTableauToFreeCell タブローからリザーブセルにカードを移動
	MoveTableauToFreeCell(col, cell int) string
	// MoveFreeCellToTableau リザーブセルからタブローにカードを移動
	MoveFreeCellToTableau(cell, col int) string
	// MoveFreeCellToFoundation リザーブセルからファンデーションにカードを移動
	MoveFreeCellToFoundation(cell int) string
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

// SeahavenTowersInteractor シーヘイブンタワーズインタラクタークラス
type SeahavenTowersInteractor struct {
	GameBase[interfaces.SeahavenTowersGame]
	sp presenter.SeahavenTowersPresenter
	solitaireActions[interfaces.SeahavenTowersGame]
}

// NewSeahavenTowersInteractor コンストラクタ
func NewSeahavenTowersInteractor(s interfaces.SeahavenTowersGame, sp presenter.SeahavenTowersPresenter) *SeahavenTowersInteractor {
	mustNotNil("SeahavenTowersInteractor", map[string]any{"s": s, "sp": sp})
	return &SeahavenTowersInteractor{
		GameBase:         GameBase[interfaces.SeahavenTowersGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.SeahavenTowersGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *SeahavenTowersInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (si *SeahavenTowersInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error {
		return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (si *SeahavenTowersInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToFreeCell タブローからリザーブセルにカードを移動
func (si *SeahavenTowersInteractor) MoveTableauToFreeCell(col, cell int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToFreeCell(col, cell) })
}

// MoveFreeCellToTableau リザーブセルからタブローにカードを移動
func (si *SeahavenTowersInteractor) MoveFreeCellToTableau(cell, col int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveFreeCellToTableau(cell, col) })
}

// MoveFreeCellToFoundation リザーブセルからファンデーションにカードを移動
func (si *SeahavenTowersInteractor) MoveFreeCellToFoundation(cell int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveFreeCellToFoundation(cell) })
}

// Hint ヒント取得
func (si *SeahavenTowersInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SeahavenTowersInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSeahavenTowersInteractor deserialises JSON into a SeahavenTowersInteractor.
func RestoreSeahavenTowersInteractor(data []byte, sp presenter.SeahavenTowersPresenter) (*SeahavenTowersInteractor, error) {
	return restoreAndBuild[domain.SeahavenTowers](data, func(g *domain.SeahavenTowers) *SeahavenTowersInteractor {
		return &SeahavenTowersInteractor{
			GameBase:         GameBase[interfaces.SeahavenTowersGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.SeahavenTowersGame](g, sp),
		}
	})
}
