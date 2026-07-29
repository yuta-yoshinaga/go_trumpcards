//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MissMilliganInteractorIF ミス・ミリガン インタラクターインタフェース
type MissMilliganInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal 山札から各列へ 1 枚ずつ配り足す
	Deal() string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(col int) string
	// Waive タブローの連番を脇へ持ち上げる
	Waive(col, cardIndex int) string
	// PlaceWaived 保持中の札をタブローへ戻す
	PlaceWaived(toCol int) string
	// MoveWaivedToFoundation 保持中の 1 枚を基礎札へ送る
	MoveWaivedToFoundation() string
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

// MissMilliganInteractor ミス・ミリガン インタラクタークラス
type MissMilliganInteractor struct {
	GameBase[interfaces.MissMilliganGame]
	mmp presenter.MissMilliganPresenter
	solitaireActions[interfaces.MissMilliganGame]
}

// NewMissMilliganInteractor コンストラクタ
func NewMissMilliganInteractor(mm interfaces.MissMilliganGame, mmp presenter.MissMilliganPresenter) *MissMilliganInteractor {
	mustNotNil("MissMilliganInteractor", map[string]any{"mm": mm, "mmp": mmp})
	return &MissMilliganInteractor{
		GameBase:         GameBase[interfaces.MissMilliganGame]{Game: mm},
		mmp:              mmp,
		solitaireActions: newSolitaireActions[interfaces.MissMilliganGame](mm, mmp),
	}
}

// Reset ゲーム初期化
func (mi *MissMilliganInteractor) Reset() string {
	return runAndPresent(mi.Game, mi.mmp, mi.Game.Reset)
}

// Deal 山札から各列へ 1 枚ずつ配り足す
func (mi *MissMilliganInteractor) Deal() string {
	return execAndPresent(mi.Game, mi.mmp, mi.Game.Deal)
}

// MoveTableauToTableau タブロー間で移動
func (mi *MissMilliganInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(mi.Game, mi.mmp, func() error {
		return mi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (mi *MissMilliganInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(mi.Game, mi.mmp, func() error { return mi.Game.MoveTableauToFoundation(col) })
}

// Waive タブローの連番を脇へ持ち上げる
func (mi *MissMilliganInteractor) Waive(col, cardIndex int) string {
	return execAndPresent(mi.Game, mi.mmp, func() error { return mi.Game.Waive(col, cardIndex) })
}

// PlaceWaived 保持中の札をタブローへ戻す
func (mi *MissMilliganInteractor) PlaceWaived(toCol int) string {
	return execAndPresent(mi.Game, mi.mmp, func() error { return mi.Game.PlaceWaived(toCol) })
}

// MoveWaivedToFoundation 保持中の 1 枚を基礎札へ送る
func (mi *MissMilliganInteractor) MoveWaivedToFoundation() string {
	return execAndPresent(mi.Game, mi.mmp, mi.Game.MoveWaivedToFoundation)
}

// Hint ヒント取得
func (mi *MissMilliganInteractor) Hint() string {
	return mi.mmp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MissMilliganInteractor) ActionLog() string {
	return mi.mmp.ActionLogOutput(mi.Game)
}

// RestoreMissMilliganInteractor deserialises JSON into a MissMilliganInteractor.
func RestoreMissMilliganInteractor(data []byte, mmp presenter.MissMilliganPresenter) (*MissMilliganInteractor, error) {
	return restoreAndBuild[domain.MissMilligan](data, func(g *domain.MissMilligan) *MissMilliganInteractor {
		return &MissMilliganInteractor{
			GameBase:         GameBase[interfaces.MissMilliganGame]{Game: g},
			mmp:              mmp,
			solitaireActions: newSolitaireActions[interfaces.MissMilliganGame](g, mmp),
		}
	})
}
