//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SomersetInteractorIF サマセットインタラクターインタフェース
type SomersetInteractorIF interface {
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

// SomersetInteractor サマセットインタラクタークラス
type SomersetInteractor struct {
	GameBase[interfaces.SomersetGame]
	bcp presenter.SomersetPresenter
	solitaireActions[interfaces.SomersetGame]
}

// NewSomersetInteractor コンストラクタ
func NewSomersetInteractor(bc interfaces.SomersetGame, bcp presenter.SomersetPresenter) *SomersetInteractor {
	mustNotNil("SomersetInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &SomersetInteractor{
		GameBase:         GameBase[interfaces.SomersetGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.SomersetGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *SomersetInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *SomersetInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *SomersetInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *SomersetInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *SomersetInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreSomersetInteractor deserialises JSON into a SomersetInteractor.
func RestoreSomersetInteractor(data []byte, bcp presenter.SomersetPresenter) (*SomersetInteractor, error) {
	return restoreAndBuild[domain.Somerset](data, func(g *domain.Somerset) *SomersetInteractor {
		return &SomersetInteractor{
			GameBase:         GameBase[interfaces.SomersetGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.SomersetGame](g, bcp),
		}
	})
}
