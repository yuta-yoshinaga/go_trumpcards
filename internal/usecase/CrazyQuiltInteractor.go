//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CrazyQuiltInteractorIF クレイジーキルト インタラクターインタフェース
type CrazyQuiltInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveQuiltToFoundation キルトの札を基礎札へ送る
	MoveQuiltToFoundation(idx int) string
	// MoveQuiltToWaste キルトの札を捨て札の上へ置く
	MoveQuiltToWaste(idx int) string
	// MoveWasteToFoundation 捨て札から基礎札へ移動
	MoveWasteToFoundation() string
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

// CrazyQuiltInteractor クレイジーキルト インタラクタークラス
type CrazyQuiltInteractor struct {
	GameBase[interfaces.CrazyQuiltGame]
	cp presenter.CrazyQuiltPresenter
	solitaireActions[interfaces.CrazyQuiltGame]
}

// NewCrazyQuiltInteractor コンストラクタ
func NewCrazyQuiltInteractor(c interfaces.CrazyQuiltGame, cp presenter.CrazyQuiltPresenter) *CrazyQuiltInteractor {
	mustNotNil("CrazyQuiltInteractor", map[string]any{"c": c, "cp": cp})
	return &CrazyQuiltInteractor{
		GameBase:         GameBase[interfaces.CrazyQuiltGame]{Game: c},
		cp:               cp,
		solitaireActions: newSolitaireActions[interfaces.CrazyQuiltGame](c, cp),
	}
}

// Reset ゲーム初期化
func (ci *CrazyQuiltInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (ci *CrazyQuiltInteractor) Draw() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Draw)
}

// MoveQuiltToFoundation キルトの札を基礎札へ送る
func (ci *CrazyQuiltInteractor) MoveQuiltToFoundation(idx int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveQuiltToFoundation(idx) })
}

// MoveQuiltToWaste キルトの札を捨て札の上へ置く
func (ci *CrazyQuiltInteractor) MoveQuiltToWaste(idx int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.MoveQuiltToWaste(idx) })
}

// MoveWasteToFoundation 捨て札から基礎札へ移動
func (ci *CrazyQuiltInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.MoveWasteToFoundation)
}

// Hint ヒント取得
func (ci *CrazyQuiltInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CrazyQuiltInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCrazyQuiltInteractor deserialises JSON into a CrazyQuiltInteractor.
func RestoreCrazyQuiltInteractor(data []byte, cp presenter.CrazyQuiltPresenter) (*CrazyQuiltInteractor, error) {
	return restoreAndBuild[domain.CrazyQuilt](data, func(g *domain.CrazyQuilt) *CrazyQuiltInteractor {
		return &CrazyQuiltInteractor{
			GameBase:         GameBase[interfaces.CrazyQuiltGame]{Game: g},
			cp:               cp,
			solitaireActions: newSolitaireActions[interfaces.CrazyQuiltGame](g, cp),
		}
	})
}
