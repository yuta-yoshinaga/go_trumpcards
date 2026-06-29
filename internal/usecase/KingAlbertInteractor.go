//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KingAlbertInteractorIF King Albert インタラクターインタフェース
type KingAlbertInteractorIF interface {
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

// KingAlbertInteractor King Albert インタラクタークラス
type KingAlbertInteractor struct {
	GameBase[interfaces.KingAlbertGame]
	bcp presenter.KingAlbertPresenter
	solitaireActions[interfaces.KingAlbertGame]
}

// NewKingAlbertInteractor コンストラクタ
func NewKingAlbertInteractor(bc interfaces.KingAlbertGame, bcp presenter.KingAlbertPresenter) *KingAlbertInteractor {
	mustNotNil("KingAlbertInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &KingAlbertInteractor{
		GameBase:         GameBase[interfaces.KingAlbertGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.KingAlbertGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *KingAlbertInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *KingAlbertInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *KingAlbertInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// MoveReserveToTableau リザーブからタブローにカードを移動
func (bi *KingAlbertInteractor) MoveReserveToTableau(reserveIdx, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveReserveToTableau(reserveIdx, toCol) })
}

// MoveReserveToFoundation リザーブからファンデーションにカードを移動
func (bi *KingAlbertInteractor) MoveReserveToFoundation(reserveIdx int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveReserveToFoundation(reserveIdx) })
}

// Hint ヒント取得
func (bi *KingAlbertInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *KingAlbertInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreKingAlbertInteractor deserialises JSON into a KingAlbertInteractor.
func RestoreKingAlbertInteractor(data []byte, bcp presenter.KingAlbertPresenter) (*KingAlbertInteractor, error) {
	return restoreAndBuild[domain.KingAlbert](data, func(g *domain.KingAlbert) *KingAlbertInteractor {
		return &KingAlbertInteractor{
			GameBase:         GameBase[interfaces.KingAlbertGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.KingAlbertGame](g, bcp),
		}
	})
}
