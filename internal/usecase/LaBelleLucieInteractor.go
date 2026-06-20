//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LaBelleLucieInteractorIF ラ・ベル・ルーシーのインタラクターインタフェース。
type LaBelleLucieInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// MoveFanToFan 扇から扇へ移す。
	MoveFanToFan(from, to int) string
	// MoveFanToFoundation 扇からファウンデーションへ移す。
	MoveFanToFoundation(from int) string
	// Redeal 集めてシャッフルし配り直す。
	Redeal() string
	// GiveUp 投了する。
	GiveUp() string
	// Hint ヒント取得。
	Hint() string
	// AutoComplete オートコンプリート。
	AutoComplete() string
	// ActionLog 棋譜出力。
	ActionLog() string
	// Undo アンドゥ。
	Undo() string
	// UndoN n 回アンドゥ。
	UndoN(n int) string
}

// LaBelleLucieInteractor ラ・ベル・ルーシーのインタラクタークラス。
type LaBelleLucieInteractor struct {
	GameBase[interfaces.LaBelleLucieGame]
	op presenter.LaBelleLuciePresenter
}

// NewLaBelleLucieInteractor コンストラクタ。
func NewLaBelleLucieInteractor(g interfaces.LaBelleLucieGame, op presenter.LaBelleLuciePresenter) *LaBelleLucieInteractor {
	mustNotNil("LaBelleLucieInteractor", map[string]any{"g": g, "op": op})
	return &LaBelleLucieInteractor{GameBase: GameBase[interfaces.LaBelleLucieGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (li *LaBelleLucieInteractor) Reset() string {
	return runAndPresent(li.Game, li.op, li.Game.Reset)
}

// MoveFanToFan 扇から扇へ移す。
func (li *LaBelleLucieInteractor) MoveFanToFan(from, to int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.MoveFanToFan(from, to) })
}

// MoveFanToFoundation 扇からファウンデーションへ移す。
func (li *LaBelleLucieInteractor) MoveFanToFoundation(from int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.MoveFanToFoundation(from) })
}

// Redeal 集めてシャッフルし配り直す。
func (li *LaBelleLucieInteractor) Redeal() string {
	return execAndPresent(li.Game, li.op, li.Game.Redeal)
}

// GiveUp 投了する。
func (li *LaBelleLucieInteractor) GiveUp() string {
	return runAndPresent(li.Game, li.op, li.Game.GiveUp)
}

// Hint ヒント取得。
func (li *LaBelleLucieInteractor) Hint() string {
	return li.op.HintOutput(li.Game)
}

// AutoComplete オートコンプリート。
func (li *LaBelleLucieInteractor) AutoComplete() string {
	return execAndPresent(li.Game, li.op, li.Game.AutoComplete)
}

// ActionLog 棋譜出力。
func (li *LaBelleLucieInteractor) ActionLog() string {
	return li.op.ActionLogOutput(li.Game)
}

// Undo アンドゥ。
func (li *LaBelleLucieInteractor) Undo() string {
	return execAndPresent(li.Game, li.op, li.Game.Undo)
}

// UndoN n 回アンドゥ。
func (li *LaBelleLucieInteractor) UndoN(n int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.UndoN(n) })
}

// RestoreLaBelleLucieInteractor deserialises JSON into a LaBelleLucieInteractor.
func RestoreLaBelleLucieInteractor(data []byte, op presenter.LaBelleLuciePresenter) (*LaBelleLucieInteractor, error) {
	return restoreAndBuild[domain.LaBelleLucie](data, func(g *domain.LaBelleLucie) *LaBelleLucieInteractor {
		return &LaBelleLucieInteractor{GameBase: GameBase[interfaces.LaBelleLucieGame]{Game: g}, op: op}
	})
}
