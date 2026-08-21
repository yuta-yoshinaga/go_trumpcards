//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NarcoticInteractorIF ナルコティックインタラクターインタフェース
type NarcoticInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 各列にカードを配る
	Draw() string
	// Remove 露出4枚のランクが揃っているとき、その4枚を捨てる
	Remove() string
	// Move 列の露出札を、同ランクを露出する最も左の列へ重ねる
	Move(col int) string
	// Redeal 山札が尽きたとき集めて配り直す
	Redeal() string
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

// NarcoticInteractor ナルコティックインタラクタークラス
type NarcoticInteractor struct {
	GameBase[interfaces.NarcoticGame]
	gp presenter.NarcoticPresenter
}

// NewNarcoticInteractor コンストラクタ
func NewNarcoticInteractor(g interfaces.NarcoticGame, gp presenter.NarcoticPresenter) *NarcoticInteractor {
	mustNotNil("NarcoticInteractor", map[string]any{"g": g, "gp": gp})
	return &NarcoticInteractor{GameBase: GameBase[interfaces.NarcoticGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ai *NarcoticInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.gp, ai.Game.Reset)
}

// Draw 各列にカードを配る
func (ai *NarcoticInteractor) Draw() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Draw)
}

// Remove 露出4枚のランクが揃っているとき、その4枚を捨てる
func (ai *NarcoticInteractor) Remove() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Remove)
}

// Redeal 山札が尽きたとき集めて配り直す
func (ai *NarcoticInteractor) Redeal() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Redeal)
}

// Move 列の一番上のカードを空き列へ移動する
func (ai *NarcoticInteractor) Move(col int) string {
	return execAndPresent(ai.Game, ai.gp, func() error { return ai.Game.Move(col) })
}

// GiveUp ギブアップ
func (ai *NarcoticInteractor) GiveUp() string {
	return runAndPresent(ai.Game, ai.gp, ai.Game.GiveUp)
}

// Hint ヒント取得
func (ai *NarcoticInteractor) Hint() string {
	return ai.gp.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *NarcoticInteractor) ActionLog() string {
	return ai.gp.ActionLogOutput(ai.Game)
}

// Undo アンドゥ
func (ai *NarcoticInteractor) Undo() string {
	return execAndPresent(ai.Game, ai.gp, ai.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ai *NarcoticInteractor) UndoN(n int) string {
	return execAndPresent(ai.Game, ai.gp, func() error { return ai.Game.UndoN(n) })
}

// RestoreNarcoticInteractor deserialises JSON into an NarcoticInteractor.
func RestoreNarcoticInteractor(data []byte, gp presenter.NarcoticPresenter) (*NarcoticInteractor, error) {
	return restoreAndBuild[domain.Narcotic](data, func(g *domain.Narcotic) *NarcoticInteractor {
		return &NarcoticInteractor{GameBase: GameBase[interfaces.NarcoticGame]{Game: g}, gp: gp}
	})
}
