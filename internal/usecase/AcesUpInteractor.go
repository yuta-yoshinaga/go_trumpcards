//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AcesUpInteractorIF エースアップインタラクターインタフェース
type AcesUpInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 各列にカードを配る
	Draw() string
	// Remove 列の一番上のカードを除去する
	Remove(col int) string
	// Move 列の一番上のカードを空き列へ移動する
	Move(col int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// AcesUpInteractor エースアップインタラクタークラス
type AcesUpInteractor struct {
	GameBase[interfaces.AcesUpGame]
	gp presenter.AcesUpPresenter
}

// NewAcesUpInteractor コンストラクタ
func NewAcesUpInteractor(g interfaces.AcesUpGame, gp presenter.AcesUpPresenter) *AcesUpInteractor {
	mustNotNil("AcesUpInteractor", map[string]any{"g": g, "gp": gp})
	return &AcesUpInteractor{GameBase: GameBase[interfaces.AcesUpGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ai *AcesUpInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.gp, ai.Game.Reset)
}

// Draw 各列にカードを配る
func (ai *AcesUpInteractor) Draw() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Draw)
}

// Remove 列の一番上のカードを除去する
func (ai *AcesUpInteractor) Remove(col int) string {
	return execAndPresent(ai.Game, ai.gp, func() error { return ai.Game.Remove(col) })
}

// Move 列の一番上のカードを空き列へ移動する
func (ai *AcesUpInteractor) Move(col int) string {
	return execAndPresent(ai.Game, ai.gp, func() error { return ai.Game.Move(col) })
}

// GiveUp ギブアップ
func (ai *AcesUpInteractor) GiveUp() string {
	return runAndPresent(ai.Game, ai.gp, ai.Game.GiveUp)
}

// Hint ヒント取得
func (ai *AcesUpInteractor) Hint() string {
	return ai.gp.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *AcesUpInteractor) ActionLog() string {
	return ai.gp.ActionLogOutput(ai.Game)
}

// Undo アンドゥ
func (ai *AcesUpInteractor) Undo() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ai *AcesUpInteractor) UndoN(n int) string {
	return execAndPresent(ai.Game, ai.gp, func() error { return ai.Game.UndoN(n) })
}

// RestoreAcesUpInteractor deserialises JSON into an AcesUpInteractor.
func RestoreAcesUpInteractor(data []byte, gp presenter.AcesUpPresenter) (*AcesUpInteractor, error) {
	return restoreAndBuild[domain.AcesUp](data, func(g *domain.AcesUp) *AcesUpInteractor {
		return &AcesUpInteractor{GameBase: GameBase[interfaces.AcesUpGame]{Game: g}, gp: gp}
	})
}
