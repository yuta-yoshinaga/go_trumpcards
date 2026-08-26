//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SlyFoxInteractorIF スライ・フォックス インタラクターインタフェース
type SlyFoxInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// DealToPile 山札から 1 枚めくって、選んだリザーブ枠に置く
	DealToPile(pile int) string
	// DealToFoundation 山札から 1 枚めくって、そのまま基礎札へ送る
	DealToFoundation(fIdx int) string
	// MoveTableauToFoundation リザーブから基礎札へ移動
	MoveTableauToFoundation(pile int) string
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

// SlyFoxInteractor スライ・フォックス インタラクタークラス
type SlyFoxInteractor struct {
	GameBase[interfaces.SlyFoxGame]
	cp presenter.SlyFoxPresenter
	solitaireActions[interfaces.SlyFoxGame]
}

// NewSlyFoxInteractor コンストラクタ
func NewSlyFoxInteractor(c interfaces.SlyFoxGame, cp presenter.SlyFoxPresenter) *SlyFoxInteractor {
	mustNotNil("SlyFoxInteractor", map[string]any{"c": c, "cp": cp})
	return &SlyFoxInteractor{
		GameBase:         GameBase[interfaces.SlyFoxGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.SlyFoxGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *SlyFoxInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// DealToPile 山札から 1 枚めくって、選んだリザーブ枠に置く
func (ci *SlyFoxInteractor) DealToPile(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.DealToPile(pile) })
}

// DealToFoundation 山札から 1 枚めくって、そのまま基礎札へ送る
func (ci *SlyFoxInteractor) DealToFoundation(fIdx int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.DealToFoundation(fIdx) })
}

// MoveTableauToFoundation リザーブから基礎札へ移動
func (ci *SlyFoxInteractor) MoveTableauToFoundation(pile int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveTableauToFoundation(pile) })
}

// Hint ヒント取得
func (ci *SlyFoxInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *SlyFoxInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreSlyFoxInteractor deserialises JSON into a SlyFoxInteractor.
func RestoreSlyFoxInteractor(data []byte, cp presenter.SlyFoxPresenter) (*SlyFoxInteractor, error) {
	return restoreAndBuild[domain.SlyFox](data, func(g *domain.SlyFox) *SlyFoxInteractor {
		return &SlyFoxInteractor{
			GameBase:         GameBase[interfaces.SlyFoxGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.SlyFoxGame](g, cp),
		}
	})
}
