//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FortyAndEightInteractorIF フォーティ・アンド・エイトインタラクターインタフェース
type FortyAndEightInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Redeal ウェイストを集めて新しいストックを作る
	Redeal() string
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

// FortyAndEightInteractor フォーティ・アンド・エイトインタラクタークラス
type FortyAndEightInteractor struct {
	GameBase[interfaces.FortyAndEightGame]
	ftp presenter.FortyAndEightPresenter
	solitaireActions[interfaces.FortyAndEightGame]
}

// NewFortyAndEightInteractor コンストラクタ
func NewFortyAndEightInteractor(ft interfaces.FortyAndEightGame, ftp presenter.FortyAndEightPresenter) *FortyAndEightInteractor {
	mustNotNil("FortyAndEightInteractor", map[string]any{"ft": ft, "ftp": ftp})
	return &FortyAndEightInteractor{
		GameBase:         GameBase[interfaces.FortyAndEightGame]{Game: ft},
		ftp:              ftp,
		solitaireActions: newSolitaireActions[interfaces.FortyAndEightGame](ft, ftp),
	}
}

// Reset ゲーム初期化
func (fi *FortyAndEightInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.ftp, fi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (fi *FortyAndEightInteractor) Draw() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.Draw)
}

// Redeal ウェイストを集めて新しいストックを作る
func (fi *FortyAndEightInteractor) Redeal() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.Redeal)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (fi *FortyAndEightInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (fi *FortyAndEightInteractor) MoveWasteToFoundation() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *FortyAndEightInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *FortyAndEightInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (fi *FortyAndEightInteractor) Hint() string {
	return fi.ftp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *FortyAndEightInteractor) ActionLog() string {
	return fi.ftp.ActionLogOutput(fi.Game)
}

// RestoreFortyAndEightInteractor deserialises JSON into a FortyAndEightInteractor.
func RestoreFortyAndEightInteractor(data []byte, ftp presenter.FortyAndEightPresenter) (*FortyAndEightInteractor, error) {
	return restoreAndBuild[domain.FortyAndEight](data, func(g *domain.FortyAndEight) *FortyAndEightInteractor {
		return &FortyAndEightInteractor{
			GameBase:         GameBase[interfaces.FortyAndEightGame]{Game: g},
			ftp:              ftp,
			solitaireActions: newSolitaireActions[interfaces.FortyAndEightGame](g, ftp),
		}
	})
}
