//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RankAndFileInteractorIF ランク・アンド・ファイルインタラクターインタフェース
type RankAndFileInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// MoveWasteToTableau ウェイストからタブローにカードを移動
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
	MoveWasteToFoundation() string
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

// RankAndFileInteractor ランク・アンド・ファイルインタラクタークラス
type RankAndFileInteractor struct {
	GameBase[interfaces.RankAndFileGame]
	ftp presenter.RankAndFilePresenter
	solitaireActions[interfaces.RankAndFileGame]
}

// NewRankAndFileInteractor コンストラクタ
func NewRankAndFileInteractor(ft interfaces.RankAndFileGame, ftp presenter.RankAndFilePresenter) *RankAndFileInteractor {
	mustNotNil("RankAndFileInteractor", map[string]any{"ft": ft, "ftp": ftp})
	return &RankAndFileInteractor{
		GameBase:         GameBase[interfaces.RankAndFileGame]{Game: ft},
		ftp:              ftp,
		solitaireActions: newSolitaireActions[interfaces.RankAndFileGame](ft, ftp),
	}
}

// Reset ゲーム初期化
func (fi *RankAndFileInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.ftp, fi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (fi *RankAndFileInteractor) Draw() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (fi *RankAndFileInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (fi *RankAndFileInteractor) MoveWasteToFoundation() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *RankAndFileInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *RankAndFileInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (fi *RankAndFileInteractor) Hint() string {
	return fi.ftp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *RankAndFileInteractor) ActionLog() string {
	return fi.ftp.ActionLogOutput(fi.Game)
}

// RestoreRankAndFileInteractor deserialises JSON into a RankAndFileInteractor.
func RestoreRankAndFileInteractor(data []byte, ftp presenter.RankAndFilePresenter) (*RankAndFileInteractor, error) {
	return restoreAndBuild[domain.RankAndFile](data, func(g *domain.RankAndFile) *RankAndFileInteractor {
		return &RankAndFileInteractor{
			GameBase:         GameBase[interfaces.RankAndFileGame]{Game: g},
			ftp:              ftp,
			solitaireActions: newSolitaireActions[interfaces.RankAndFileGame](g, ftp),
		}
	})
}
