//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BeleagueredCastleInteractorIF Beleaguered Castle インタラクターインタフェース
type BeleagueredCastleInteractorIF interface {
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

// BeleagueredCastleInteractor Beleaguered Castle インタラクタークラス
type BeleagueredCastleInteractor struct {
	GameBase[interfaces.BeleagueredCastleGame]
	bcp presenter.BeleagueredCastlePresenter
	solitaireActions[interfaces.BeleagueredCastleGame]
}

// NewBeleagueredCastleInteractor コンストラクタ
func NewBeleagueredCastleInteractor(bc interfaces.BeleagueredCastleGame, bcp presenter.BeleagueredCastlePresenter) *BeleagueredCastleInteractor {
	mustNotNil("BeleagueredCastleInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &BeleagueredCastleInteractor{
		GameBase:         GameBase[interfaces.BeleagueredCastleGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.BeleagueredCastleGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *BeleagueredCastleInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *BeleagueredCastleInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *BeleagueredCastleInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *BeleagueredCastleInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BeleagueredCastleInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreBeleagueredCastleInteractor deserialises JSON into a BeleagueredCastleInteractor.
func RestoreBeleagueredCastleInteractor(data []byte, bcp presenter.BeleagueredCastlePresenter) (*BeleagueredCastleInteractor, error) {
	return restoreAndBuild[domain.BeleagueredCastle](data, func(g *domain.BeleagueredCastle) *BeleagueredCastleInteractor {
		return &BeleagueredCastleInteractor{
			GameBase:         GameBase[interfaces.BeleagueredCastleGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.BeleagueredCastleGame](g, bcp),
		}
	})
}
