//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShamrocksInteractorIF シャムロックスのインタラクターインタフェース。
type ShamrocksInteractorIF interface {
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

// ShamrocksInteractor シャムロックスのインタラクタークラス。
type ShamrocksInteractor struct {
	GameBase[interfaces.ShamrocksGame]
	op presenter.ShamrocksPresenter
}

// NewShamrocksInteractor コンストラクタ。
func NewShamrocksInteractor(g interfaces.ShamrocksGame, op presenter.ShamrocksPresenter) *ShamrocksInteractor {
	mustNotNil("ShamrocksInteractor", map[string]any{"g": g, "op": op})
	return &ShamrocksInteractor{GameBase: GameBase[interfaces.ShamrocksGame]{Game: g}, op: op}
}

// Reset ゲーム初期化。
func (li *ShamrocksInteractor) Reset() string {
	return runAndPresent(li.Game, li.op, li.Game.Reset)
}

// MoveFanToFan 扇から扇へ移す。
func (li *ShamrocksInteractor) MoveFanToFan(from, to int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.MoveFanToFan(from, to) })
}

// MoveFanToFoundation 扇からファウンデーションへ移す。
func (li *ShamrocksInteractor) MoveFanToFoundation(from int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.MoveFanToFoundation(from) })
}

// Redeal 集めてシャッフルし配り直す。
func (li *ShamrocksInteractor) Redeal() string {
	return execAndPresent(li.Game, li.op, li.Game.Redeal)
}

// GiveUp 投了する。
func (li *ShamrocksInteractor) GiveUp() string {
	return runAndPresent(li.Game, li.op, li.Game.GiveUp)
}

// Hint ヒント取得。
func (li *ShamrocksInteractor) Hint() string {
	return li.op.HintOutput(li.Game)
}

// AutoComplete オートコンプリート。
func (li *ShamrocksInteractor) AutoComplete() string {
	return execAndPresent(li.Game, li.op, li.Game.AutoComplete)
}

// ActionLog 棋譜出力。
func (li *ShamrocksInteractor) ActionLog() string {
	return li.op.ActionLogOutput(li.Game)
}

// Undo アンドゥ。
func (li *ShamrocksInteractor) Undo() string {
	return execAndPresent(li.Game, li.op, li.Game.Undo)
}

// UndoN n 回アンドゥ。
func (li *ShamrocksInteractor) UndoN(n int) string {
	return execAndPresent(li.Game, li.op, func() error { return li.Game.UndoN(n) })
}

// RestoreShamrocksInteractor deserialises JSON into a ShamrocksInteractor.
func RestoreShamrocksInteractor(data []byte, op presenter.ShamrocksPresenter) (*ShamrocksInteractor, error) {
	return restoreAndBuild[domain.Shamrocks](data, func(g *domain.Shamrocks) *ShamrocksInteractor {
		return &ShamrocksInteractor{GameBase: GameBase[interfaces.ShamrocksGame]{Game: g}, op: op}
	})
}
