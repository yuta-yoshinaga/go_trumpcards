package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FortyThievesInteractorIF フォーティシーブスインタラクターインタフェース
type FortyThievesInteractorIF interface {
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

// FortyThievesInteractor フォーティシーブスインタラクタークラス
type FortyThievesInteractor struct {
	GameBase[interfaces.FortyThievesGame]
	ftp presenter.FortyThievesPresenter
	solitaireActions[interfaces.FortyThievesGame]
}

// NewFortyThievesInteractor コンストラクタ
func NewFortyThievesInteractor(ft interfaces.FortyThievesGame, ftp presenter.FortyThievesPresenter) *FortyThievesInteractor {
	mustNotNil("FortyThievesInteractor", map[string]any{"ft": ft, "ftp": ftp})
	return &FortyThievesInteractor{
		GameBase:         GameBase[interfaces.FortyThievesGame]{Game: ft},
		ftp:              ftp,
		solitaireActions: newSolitaireActions[interfaces.FortyThievesGame](ft, ftp),
	}
}

// Reset ゲーム初期化
func (fi *FortyThievesInteractor) Reset() string {
	return runAndPresent(fi.Game, fi.ftp, fi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (fi *FortyThievesInteractor) Draw() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (fi *FortyThievesInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (fi *FortyThievesInteractor) MoveWasteToFoundation() string {
	return execAndPresent(fi.Game, fi.ftp, fi.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (fi *FortyThievesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fi *FortyThievesInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(fi.Game, fi.ftp, func() error { return fi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (fi *FortyThievesInteractor) Hint() string {
	return fi.ftp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *FortyThievesInteractor) ActionLog() string {
	return fi.ftp.ActionLogOutput(fi.Game)
}

// RestoreFortyThievesInteractor deserialises JSON into a FortyThievesInteractor.
func RestoreFortyThievesInteractor(data []byte, ftp presenter.FortyThievesPresenter) (*FortyThievesInteractor, error) {
	return restoreAndBuild[domain.FortyThieves](data, func(g *domain.FortyThieves) *FortyThievesInteractor {
		return &FortyThievesInteractor{
			GameBase:         GameBase[interfaces.FortyThievesGame]{Game: g},
			ftp:              ftp,
			solitaireActions: newSolitaireActions[interfaces.FortyThievesGame](g, ftp),
		}
	})
}
