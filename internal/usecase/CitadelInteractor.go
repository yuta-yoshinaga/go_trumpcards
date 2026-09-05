//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CitadelInteractorIF Citadel インタラクターインタフェース
type CitadelInteractorIF interface {
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

// CitadelInteractor Citadel インタラクタークラス
type CitadelInteractor struct {
	GameBase[interfaces.CitadelGame]
	bcp presenter.CitadelPresenter
	solitaireActions[interfaces.CitadelGame]
}

// NewCitadelInteractor コンストラクタ
func NewCitadelInteractor(bc interfaces.CitadelGame, bcp presenter.CitadelPresenter) *CitadelInteractor {
	mustNotNil("CitadelInteractor", map[string]any{"bc": bc, "bcp": bcp})
	return &CitadelInteractor{
		GameBase:         GameBase[interfaces.CitadelGame]{Game: bc},
		bcp:              bcp,
		solitaireActions: newSolitaireActions[interfaces.CitadelGame](bc, bcp),
	}
}

// Reset ゲーム初期化
func (bi *CitadelInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bcp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *CitadelInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *CitadelInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bcp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (bi *CitadelInteractor) Hint() string {
	return bi.bcp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *CitadelInteractor) ActionLog() string {
	return bi.bcp.ActionLogOutput(bi.Game)
}

// RestoreCitadelInteractor deserialises JSON into a CitadelInteractor.
func RestoreCitadelInteractor(data []byte, bcp presenter.CitadelPresenter) (*CitadelInteractor, error) {
	return restoreAndBuild[domain.Citadel](data, func(g *domain.Citadel) *CitadelInteractor {
		return &CitadelInteractor{
			GameBase:         GameBase[interfaces.CitadelGame]{Game: g},
			bcp:              bcp,
			solitaireActions: newSolitaireActions[interfaces.CitadelGame](g, bcp),
		}
	})
}
