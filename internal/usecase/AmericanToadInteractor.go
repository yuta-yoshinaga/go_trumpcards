//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AmericanToadInteractorIF アメリカン・トード インタラクターインタフェース
type AmericanToadInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveReserveToFoundation リザーブから基礎札へ移動
	MoveReserveToFoundation() string
	// MoveReserveToTableau リザーブからタブローへ移動
	MoveReserveToTableau(col int) string
	// MoveWasteToFoundation 捨て札から基礎札へ移動
	MoveWasteToFoundation() string
	// MoveWasteToTableau 捨て札からタブローへ移動
	MoveWasteToTableau(col int) string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(col int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
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

// AmericanToadInteractor アメリカン・トード インタラクタークラス
type AmericanToadInteractor struct {
	GameBase[interfaces.AmericanToadGame]
	ap presenter.AmericanToadPresenter
	solitaireActions[interfaces.AmericanToadGame]
}

// NewAmericanToadInteractor コンストラクタ
func NewAmericanToadInteractor(at interfaces.AmericanToadGame, ap presenter.AmericanToadPresenter) *AmericanToadInteractor {
	mustNotNil("AmericanToadInteractor", map[string]any{"at": at, "ap": ap})
	return &AmericanToadInteractor{
		GameBase:         GameBase[interfaces.AmericanToadGame]{Game: at},
		ap:               ap,
		solitaireActions: newSolitaireActions[interfaces.AmericanToadGame](at, ap),
	}
}

// Reset ゲーム初期化
func (ai *AmericanToadInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.ap, ai.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ai *AmericanToadInteractor) Draw() string {
	return execAndPresent(ai.Game, ai.ap, ai.Game.Draw)
}

// MoveReserveToFoundation リザーブから基礎札へ移動
func (ai *AmericanToadInteractor) MoveReserveToFoundation() string {
	return execAndPresent(ai.Game, ai.ap, ai.Game.MoveReserveToFoundation)
}

// MoveReserveToTableau リザーブからタブローへ移動
func (ai *AmericanToadInteractor) MoveReserveToTableau(col int) string {
	return execAndPresent(ai.Game, ai.ap, func() error { return ai.Game.MoveReserveToTableau(col) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ai *AmericanToadInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ai.Game, ai.ap, ai.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau 捨て札からタブローへ移動
func (ai *AmericanToadInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(ai.Game, ai.ap, func() error { return ai.Game.MoveWasteToTableau(col) })
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (ai *AmericanToadInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ai.Game, ai.ap, func() error { return ai.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToTableau タブロー間で移動
func (ai *AmericanToadInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ai.Game, ai.ap, func() error {
		return ai.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// Hint ヒント取得
func (ai *AmericanToadInteractor) Hint() string {
	return ai.ap.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *AmericanToadInteractor) ActionLog() string {
	return ai.ap.ActionLogOutput(ai.Game)
}

// RestoreAmericanToadInteractor deserialises JSON into an AmericanToadInteractor.
func RestoreAmericanToadInteractor(data []byte, ap presenter.AmericanToadPresenter) (*AmericanToadInteractor, error) {
	return restoreAndBuild[domain.AmericanToad](data, func(g *domain.AmericanToad) *AmericanToadInteractor {
		return &AmericanToadInteractor{
			GameBase:         GameBase[interfaces.AmericanToadGame]{Game: g},
			ap:               ap,
			solitaireActions: newSolitaireActions[interfaces.AmericanToadGame](g, ap),
		}
	})
}
