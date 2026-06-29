//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AgnesInteractorIF アグネス・ソレルインタラクターインタフェース
type AgnesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// DealStock ストックから各列に1枚ずつ配る
	DealStock() string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションに移動
	MoveTableauToFoundation(col int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜出力
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// AgnesInteractor アグネス・ソレルインタラクタークラス
type AgnesInteractor struct {
	GameBase[interfaces.AgnesGame]
	cp presenter.AgnesPresenter
}

// NewAgnesInteractor コンストラクタ
func NewAgnesInteractor(c interfaces.AgnesGame, cp presenter.AgnesPresenter) *AgnesInteractor {
	mustNotNil("AgnesInteractor", map[string]any{"c": c, "cp": cp})
	return &AgnesInteractor{GameBase: GameBase[interfaces.AgnesGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ai *AgnesInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.cp, ai.Game.Reset)
}

// DealStock ストックから各列に1枚ずつ配る
func (ai *AgnesInteractor) DealStock() string {
	return execAndPresent(ai.Game, ai.cp, ai.Game.DealStock)
}

// MoveTableauToTableau タブロー間で移動
func (ai *AgnesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ai.Game, ai.cp, func() error { return ai.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションに移動
func (ai *AgnesInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ai.Game, ai.cp, func() error { return ai.Game.MoveTableauToFoundation(col) })
}

// GiveUp ギブアップ
func (ai *AgnesInteractor) GiveUp() string {
	return runAndPresent(ai.Game, ai.cp, ai.Game.GiveUp)
}

// Hint ヒント取得
func (ai *AgnesInteractor) Hint() string {
	return ai.cp.HintOutput(ai.Game)
}

// ActionLog 棋譜出力
func (ai *AgnesInteractor) ActionLog() string {
	return ai.cp.ActionLogOutput(ai.Game)
}

// Undo アンドゥ
func (ai *AgnesInteractor) Undo() string {
	return execAndPresent(ai.Game, ai.cp, ai.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ai *AgnesInteractor) UndoN(n int) string {
	return execAndPresent(ai.Game, ai.cp, func() error { return ai.Game.UndoN(n) })
}

// RestoreAgnesInteractor deserialises JSON into an AgnesInteractor.
func RestoreAgnesInteractor(data []byte, cp presenter.AgnesPresenter) (*AgnesInteractor, error) {
	return restoreAndBuild[domain.Agnes](data, func(g *domain.Agnes) *AgnesInteractor {
		return &AgnesInteractor{GameBase: GameBase[interfaces.AgnesGame]{Game: g}, cp: cp}
	})
}
