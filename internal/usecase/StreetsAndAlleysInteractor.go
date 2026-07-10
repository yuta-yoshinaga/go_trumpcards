//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// StreetsAndAlleysInteractorIF Streets and Alleys インタラクターインタフェース
type StreetsAndAlleysInteractorIF interface {
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

// StreetsAndAlleysInteractor Streets and Alleys インタラクタークラス
type StreetsAndAlleysInteractor struct {
	GameBase[interfaces.StreetsAndAlleysGame]
	bcp presenter.StreetsAndAlleysPresenter
	solitaireActions[interfaces.StreetsAndAlleysGame]
}

// NewStreetsAndAlleysInteractor コンストラクタ
func NewStreetsAndAlleysInteractor(bc interfaces.StreetsAndAlleysGame, bcp presenter.StreetsAndAlleysPresenter) *StreetsAndAlleysInteractor {
	mustNotNil("StreetsAndAlleysInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &StreetsAndAlleysInteractor{
		GameBase:         GameBase[interfaces.StreetsAndAlleysGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.StreetsAndAlleysGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *StreetsAndAlleysInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *StreetsAndAlleysInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *StreetsAndAlleysInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *StreetsAndAlleysInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *StreetsAndAlleysInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreStreetsAndAlleysInteractor deserialises JSON into a StreetsAndAlleysInteractor.
func RestoreStreetsAndAlleysInteractor(data []byte, bcp presenter.StreetsAndAlleysPresenter) (*StreetsAndAlleysInteractor, error) {
	return restoreAndBuild[domain.StreetsAndAlleys](data, func(g *domain.StreetsAndAlleys) *StreetsAndAlleysInteractor {
		return &StreetsAndAlleysInteractor{
			GameBase:         GameBase[interfaces.StreetsAndAlleysGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.StreetsAndAlleysGame](g, bcp),
		}
	})
}
