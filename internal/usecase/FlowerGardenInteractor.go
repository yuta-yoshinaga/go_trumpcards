//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FlowerGardenInteractorIF Flower Garden インタラクターインタフェース
type FlowerGardenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// MoveReserveToTableau リザーブからタブローにカードを移動
	MoveReserveToTableau(reserveIdx, toCol int) string
	// MoveReserveToFoundation リザーブからファンデーションにカードを移動
	MoveReserveToFoundation(reserveIdx int) string
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

// FlowerGardenInteractor Flower Garden インタラクタークラス
type FlowerGardenInteractor struct {
	GameBase[interfaces.FlowerGardenGame]
	bcp presenter.FlowerGardenPresenter
	solitaireActions[interfaces.FlowerGardenGame]
}

// NewFlowerGardenInteractor コンストラクタ
func NewFlowerGardenInteractor(bc interfaces.FlowerGardenGame, bcp presenter.FlowerGardenPresenter) *FlowerGardenInteractor {
	mustNotNil("FlowerGardenInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &FlowerGardenInteractor{
		GameBase:         GameBase[interfaces.FlowerGardenGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.FlowerGardenGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *FlowerGardenInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *FlowerGardenInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *FlowerGardenInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// MoveReserveToTableau リザーブからタブローにカードを移動
func (bi *FlowerGardenInteractor) MoveReserveToTableau(reserveIdx, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveReserveToTableau(reserveIdx, toCol) })
}

// MoveReserveToFoundation リザーブからファンデーションにカードを移動
func (bi *FlowerGardenInteractor) MoveReserveToFoundation(reserveIdx int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveReserveToFoundation(reserveIdx) })
}

// Hint ヒント取得
func (bi *FlowerGardenInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *FlowerGardenInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreFlowerGardenInteractor deserialises JSON into a FlowerGardenInteractor.
func RestoreFlowerGardenInteractor(data []byte, bcp presenter.FlowerGardenPresenter) (*FlowerGardenInteractor, error) {
	return restoreAndBuild[domain.FlowerGarden](data, func(g *domain.FlowerGarden) *FlowerGardenInteractor {
		return &FlowerGardenInteractor{
			GameBase:         GameBase[interfaces.FlowerGardenGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.FlowerGardenGame](g, bcp),
		}
	})
}
