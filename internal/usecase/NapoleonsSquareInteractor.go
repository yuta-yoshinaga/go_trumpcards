//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NapoleonsSquareInteractorIF ナポレオンズ・スクエア インタラクターインタフェース
type NapoleonsSquareInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札からウェイストへ 1 枚めくる
	Draw() string
	// MoveWasteToTableau ウェイストからタブローへ移動
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストから基礎札へ移動
	MoveWasteToFoundation() string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローから基礎札へ移動
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

// NapoleonsSquareInteractor ナポレオンズ・スクエア インタラクタークラス
type NapoleonsSquareInteractor struct {
	GameBase[interfaces.NapoleonsSquareGame]
	nsp presenter.NapoleonsSquarePresenter
	solitaireActions[interfaces.NapoleonsSquareGame]
}

// NewNapoleonsSquareInteractor コンストラクタ
func NewNapoleonsSquareInteractor(ns interfaces.NapoleonsSquareGame, nsp presenter.NapoleonsSquarePresenter) *NapoleonsSquareInteractor {
	mustNotNil("NapoleonsSquareInteractor", map[string]any{"ns": ns, "nsp": nsp})
	return &NapoleonsSquareInteractor{
		GameBase:         GameBase[interfaces.NapoleonsSquareGame]{Game: ns},
		nsp:              nsp,
		solitaireActions: newSolitaireActions[interfaces.NapoleonsSquareGame](ns, nsp),
	}
}

// Reset ゲーム初期化
func (ni *NapoleonsSquareInteractor) Reset() string {
	return runAndPresent(ni.Game, ni.nsp, ni.Game.Reset)
}

// Draw 山札からウェイストへ 1 枚めくる
func (ni *NapoleonsSquareInteractor) Draw() string {
	return execAndPresent(ni.Game, ni.nsp, ni.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローへ移動
func (ni *NapoleonsSquareInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(ni.Game, ni.nsp, func() error { return ni.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストから基礎札へ移動
func (ni *NapoleonsSquareInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ni.Game, ni.nsp, ni.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブロー間で移動
func (ni *NapoleonsSquareInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ni.Game, ni.nsp, func() error {
		return ni.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ni *NapoleonsSquareInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ni.Game, ni.nsp, func() error { return ni.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (ni *NapoleonsSquareInteractor) Hint() string {
	return ni.nsp.HintOutput(ni.Game)
}

// ActionLog 棋譜を出力する
func (ni *NapoleonsSquareInteractor) ActionLog() string {
	return ni.nsp.ActionLogOutput(ni.Game)
}

// RestoreNapoleonsSquareInteractor deserialises JSON into a NapoleonsSquareInteractor.
func RestoreNapoleonsSquareInteractor(data []byte, nsp presenter.NapoleonsSquarePresenter) (*NapoleonsSquareInteractor, error) {
	return restoreAndBuild[domain.NapoleonsSquare](data, func(g *domain.NapoleonsSquare) *NapoleonsSquareInteractor {
		return &NapoleonsSquareInteractor{
			GameBase:         GameBase[interfaces.NapoleonsSquareGame]{Game: g},
			nsp:              nsp,
			solitaireActions: newSolitaireActions[interfaces.NapoleonsSquareGame](g, nsp),
		}
	})
}
